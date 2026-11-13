package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"be-presensi/internal/modules/notification/entity"
	"be-presensi/internal/modules/notification/provider/whatsapp"
	notificationRepository "be-presensi/internal/modules/notification/repository"
	userRepository "be-presensi/internal/modules/user/repository"
)

var (
	ErrWhatsAppConfigNotFound  = errors.New("whatsapp configuration not found")
	ErrWhatsAppTokenEmpty      = errors.New("whatsapp api token is empty")
	ErrWhatsAppProviderInvalid = errors.New("unsupported whatsapp provider")
)

type NotificationService interface {
	SendWhatsAppMessage(
		ctx context.Context,
		target string,
		message string,
	) (*whatsapp.SendMessageResponse, error)

	SendWhatsAppToAdmins(
		ctx context.Context,
		attendanceID int64,
		userID int64,
		message string,
	) error

	SendWhatsAppToUser(
		ctx context.Context,
		userID int64,
		target string,
		message string,
	) error
}

type notificationService struct {
	repository     notificationRepository.WhatsAppConfigRepository
	logRepository  notificationRepository.WhatsAppLogRepository
	userRepository userRepository.UserRepository
	whatsappClient *whatsapp.WhatsAppClient
}

type whatsappAttemptDetail struct {
	ConfigID  int64                         `json:"config_id"`
	Priority  int                           `json:"priority"`
	Status    string                        `json:"status"`
	Error     string                        `json:"error,omitempty"`
	Response  *whatsapp.SendMessageResponse `json:"response,omitempty"`
	AttemptAt time.Time                     `json:"attempt_at"`
}

func NewNotificationService(
	repository notificationRepository.WhatsAppConfigRepository,
	logRepository notificationRepository.WhatsAppLogRepository,
	userRepository userRepository.UserRepository,
	whatsappClient *whatsapp.WhatsAppClient,
) NotificationService {
	return &notificationService{
		repository:     repository,
		logRepository:  logRepository,
		userRepository: userRepository,
		whatsappClient: whatsappClient,
	}
}

func (s *notificationService) SendWhatsAppMessage(
	ctx context.Context,
	target string,
	message string,
) (*whatsapp.SendMessageResponse, error) {
	return s.sendWhatsAppMessage(ctx, target, message, nil, nil)
}

func (s *notificationService) sendWhatsAppMessage(
	ctx context.Context,
	target string,
	message string,
	attendanceID *int64,
	userID *int64,
) (*whatsapp.SendMessageResponse, error) {
	target = strings.TrimSpace(target)
	message = strings.TrimSpace(message)

	if target == "" || message == "" {
		return nil, errors.New("whatsapp target and message are required")
	}

	configs, err := s.repository.FindAllActive()
	if err != nil {
		return nil, err
	}

	if len(configs) == 0 {
		return nil, ErrWhatsAppConfigNotFound
	}

	var (
		lastErr     error
		result      *whatsapp.SendMessageResponse
		successID   *int64
		attempts    []*whatsappAttemptDetail
		finalStatus = "FAILED"
	)

	for _, config := range configs {
		attempt := &whatsappAttemptDetail{
			ConfigID:  config.ID,
			Priority:  config.Priority,
			AttemptAt: time.Now(),
		}

		// Provider tidak didukung.
		if !strings.EqualFold(config.Provider, "fonnte") {
			lastErr = ErrWhatsAppProviderInvalid
			attempt.Status = "FAILED"
			attempt.Error = lastErr.Error()
			attempts = append(attempts, attempt)
			continue
		}

		// Token kosong.
		if strings.TrimSpace(config.APIToken) == "" {
			lastErr = ErrWhatsAppTokenEmpty
			attempt.Status = "FAILED"
			attempt.Error = lastErr.Error()
			attempts = append(attempts, attempt)
			continue
		}

		response, sendErr := s.whatsappClient.SendMessage(
			ctx,
			config.APIToken,
			&whatsapp.SendMessageRequest{
				Target:  target,
				Message: message,
			},
		)

		// Pengiriman berhasil.
		if sendErr == nil {
			attempt.Status = "SENT"
			attempt.Response = response
			attempts = append(attempts, attempt)

			result = response

			configID := config.ID
			successID = &configID

			lastErr = nil
			finalStatus = "SENT"
			break
		}

		lastErr = fmt.Errorf(
			"fonnte config ID %d failed: %w",
			config.ID,
			sendErr,
		)

		attempt.Error = sendErr.Error()
		attempt.Response = response

		// Fonnte secara eksplisit menolak pesan.
		// Boleh mencoba token berikutnya.
		if whatsapp.IsProviderRejected(sendErr) {
			attempt.Status = "FAILED"
			attempts = append(attempts, attempt)
			continue
		}

		// Status pengiriman tidak pasti.
		// Hentikan fallback untuk mencegah pesan ganda.
		attempt.Status = "UNCERTAIN"
		attempts = append(attempts, attempt)

		finalStatus = "UNCERTAIN"
		break
	}

	// Persiapkan informasi log.
	status := finalStatus

	var errorMessage *string
	var sentAt *time.Time

	if lastErr != nil {
		errText := lastErr.Error()
		errorMessage = &errText
	} else {
		now := time.Now()
		sentAt = &now
	}

	// Serialize detail percobaan.
	attemptJSON, marshalErr := json.Marshal(attempts)
	if marshalErr != nil {
		return result, marshalErr
	}

	attemptDetails := string(attemptJSON)

	// Simpan log WhatsApp.
	whatsappLog := &entity.WhatsAppLog{
		AttendanceID:       attendanceID,
		UserID:             userID,
		Provider:           "fonnte",
		Target:             target,
		Message:            message,
		Status:             status,
		AttemptCount:       len(attempts),
		SuccessfulConfigID: successID,
		AttemptDetails:     &attemptDetails,
		ErrorMessage:       errorMessage,
		SentAt:             sentAt,
	}

	if logErr := s.logRepository.Create(ctx, whatsappLog); logErr != nil {
		log.Printf(
			"WARNING: failed to save WhatsApp log. AttendanceID=%v, Target=%s, Error=%v",
			attendanceID,
			target,
			logErr,
		)
	}

	// Kembalikan error sesuai kondisi pengiriman.
	if lastErr != nil {
		if finalStatus == "UNCERTAIN" {
			return nil, fmt.Errorf(
				"whatsapp delivery status is uncertain: %w",
				lastErr,
			)
		}

		return nil, fmt.Errorf(
			"all whatsapp configurations failed: %w",
			lastErr,
		)
	}

	return result, nil
}

func (s *notificationService) SendWhatsAppToAdmins(
	ctx context.Context,
	attendanceID int64,
	userID int64,
	message string,
) error {
	message = strings.TrimSpace(message)
	if message == "" {
		return errors.New("whatsapp message is required")
	}

	phoneNumbers, err := s.userRepository.FindActiveAdminPhoneNumbers()
	if err != nil {
		return err
	}

	if len(phoneNumbers) == 0 {
		return errors.New("active admin phone number not found")
	}

	var sendErrors []string

	for _, phoneNumber := range phoneNumbers {
		_, err := s.sendWhatsAppMessage(
			ctx,
			phoneNumber,
			message,
			&attendanceID,
			&userID,
		)
		if err != nil {
			sendErrors = append(
				sendErrors,
				fmt.Sprintf("%s: %v", phoneNumber, err),
			)
		}
	}

	if len(sendErrors) > 0 {
		return fmt.Errorf(
			"failed to send whatsapp to some admins: %s",
			strings.Join(sendErrors, "; "),
		)
	}

	return nil
}

func (s *notificationService) SendWhatsAppToUser(
	ctx context.Context,
	userID int64,
	target string,
	message string,
) error {
	target = strings.TrimSpace(target)
	message = strings.TrimSpace(message)

	if userID <= 0 {
		return errors.New("invalid whatsapp recipient user ID")
	}

	if target == "" || message == "" {
		return errors.New("whatsapp target and message are required")
	}

	_, err := s.sendWhatsAppMessage(
		ctx,
		target,
		message,
		nil,
		&userID,
	)

	return err
}
