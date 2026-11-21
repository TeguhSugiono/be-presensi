package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"be-presensi/internal/config"
	"be-presensi/internal/modules/activity_comment/dto"
	"be-presensi/internal/modules/activity_comment/entity"
	activityCommentRepository "be-presensi/internal/modules/activity_comment/repository"
	dailyActivityRepository "be-presensi/internal/modules/daily_activity/repository"
	notificationService "be-presensi/internal/modules/notification/service"
	userRepository "be-presensi/internal/modules/user/repository"

	"gorm.io/gorm"
)

var (
	ErrActivityCommentForbidden     = errors.New("only Admin can create activity comments")
	ErrActivityCommentRequired      = errors.New("comment is required")
	ErrDailyActivityNotFound        = errors.New("daily activity not found")
	ErrCommentOwnerNotFound         = errors.New("daily activity owner not found")
	ErrInvalidActivityCommentStatus = errors.New("invalid activity comment status")
	ErrDailyActivityNotSubmitted    = errors.New("daily activity is not submitted")
)

type ActivityCommentService interface {
	Create(
		ctx context.Context,
		kepsekID int64,
		req *dto.CreateActivityCommentRequest,
	) (*entity.ActivityComment, error)

	GetByDailyActivityID(
		ctx context.Context,
		userID int64,
		dailyActivityID int64,
	) ([]*entity.ActivityComment, error)
}

type activityCommentService struct {
	db                *gorm.DB
	cfg               *config.Config
	commentRepo       activityCommentRepository.ActivityCommentRepository
	dailyActivityRepo dailyActivityRepository.DailyActivityRepository
	userRepo          userRepository.UserRepository
	notificationSvc   notificationService.NotificationService
}

func NewActivityCommentService(
	db *gorm.DB,
	cfg *config.Config,
	commentRepo activityCommentRepository.ActivityCommentRepository,
	dailyActivityRepo dailyActivityRepository.DailyActivityRepository,
	userRepo userRepository.UserRepository,
	notificationSvc notificationService.NotificationService,
) ActivityCommentService {
	return &activityCommentService{
		db:                db,
		cfg:               cfg,
		commentRepo:       commentRepo,
		dailyActivityRepo: dailyActivityRepo,
		userRepo:          userRepo,
		notificationSvc:   notificationSvc,
	}
}

func (s *activityCommentService) Create(
	ctx context.Context,
	kepsekID int64,
	req *dto.CreateActivityCommentRequest,
) (*entity.ActivityComment, error) {

	if kepsekID <= 0 {
		return nil, ErrActivityCommentForbidden
	}

	if req.DailyActivityID <= 0 {
		return nil, ErrDailyActivityNotFound
	}

	commentText := strings.TrimSpace(req.Comment)
	if commentText == "" {
		return nil, ErrActivityCommentRequired
	}

	if req.Status != 1 && req.Status != 2 {
		return nil, ErrInvalidActivityCommentStatus
	}

	// Validate Admin
	admin, err := s.userRepo.FindByIDWithRole(kepsekID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrActivityCommentForbidden
		}
		return nil, err
	}

	if admin.Role == nil ||
		!strings.EqualFold(strings.TrimSpace(*admin.Role), "Admin") {
		return nil, ErrActivityCommentForbidden
	}

	// Find Daily Activity and owner
	activity, err := s.dailyActivityRepo.FindByID(
		ctx,
		req.DailyActivityID,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDailyActivityNotFound
		}
		return nil, err
	}

	owner, err := s.userRepo.FindByID(activity.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentOwnerNotFound
		}
		return nil, err
	}

	now := time.Now()

	var comment *entity.ActivityComment

	// Transaction: update Daily Activity and create comment together.
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		txActivityRepo := s.dailyActivityRepo.WithTx(tx)
		txCommentRepo := s.commentRepo.WithTx(tx)

		// Lock Daily Activity row
		lockedActivity, err := txActivityRepo.FindByIDForUpdate(
			ctx,
			req.DailyActivityID,
		)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrDailyActivityNotFound
			}
			return err
		}

		// Only submitted activities can be reviewed.
		if lockedActivity.Status != 2 {
			return ErrDailyActivityNotSubmitted
		}

		// Map comment decision to Daily Activity status.
		if req.Status == 1 {
			// Approved
			lockedActivity.Status = 3
			lockedActivity.IsLocked = true
		} else {
			// Revision
			lockedActivity.Status = 4
			lockedActivity.IsLocked = false
		}

		lockedActivity.UpdatedBy = &kepsekID
		lockedActivity.UpdatedAt = &now

		// Update Daily Activity status
		if err := txActivityRepo.UpdateReviewStatus(
			ctx,
			lockedActivity,
		); err != nil {
			return err
		}

		// Create Activity Comment
		comment = &entity.ActivityComment{
			DailyActivityID: req.DailyActivityID,
			KepsekID:        kepsekID,
			Comment:         commentText,
			Status:          req.Status,
			WANotified:      false,
			CreatedBy:       &kepsekID,
			CreatedAt:       now,
			UpdatedBy:       &kepsekID,
			UpdatedAt:       now,
			IsDeleted:       false,
		}

		if err := txCommentRepo.Create(ctx, comment); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Set tanggal aktivitas untuk response setelah transaksi berhasil.
	comment.DailyTaskDate = activity.ActivityDate

	// WhatsApp notification disabled
	if !s.cfg.ActivityCommentWhatsAppEnabled {
		return comment, nil
	}

	// Skip notification for inactive users or empty phone numbers.
	if !owner.IsActive || strings.TrimSpace(owner.PhoneNumber) == "" {
		return comment, nil
	}

	activityDate := activity.ActivityDate.Format("02-01-2006")

	message := fmt.Sprintf(
		"Halo %s,\n\nKepala Sekolah memberikan komentar pada aktivitas harian Anda tanggal %s:\n\n%s\n\nSilakan buka aplikasi Presensi untuk melihat detail aktivitas.",
		owner.FullName,
		activityDate,
		commentText,
	)

	// Send WhatsApp notification after transaction committed.
	if err := s.notificationSvc.SendWhatsAppToUser(
		ctx,
		owner.ID,
		owner.PhoneNumber,
		message,
	); err != nil {
		log.Printf(
			"WARNING: failed to send activity comment WhatsApp. comment_id=%d user_id=%d error=%v",
			comment.ID,
			owner.ID,
			err,
		)

		return comment, nil
	}

	// Update notification status after successful delivery.
	if err := s.commentRepo.UpdateWANotified(
		ctx,
		comment.ID,
		true,
	); err != nil {
		log.Printf(
			"WARNING: WhatsApp sent but failed to update wa_notified. comment_id=%d error=%v",
			comment.ID,
			err,
		)

		return comment, nil
	}

	comment.WANotified = true

	return comment, nil
}

// func (s *activityCommentService) Create(
// 	ctx context.Context,
// 	kepsekID int64,
// 	req dto.CreateActivityCommentRequest,
// ) (*entity.ActivityComment, error) {

// 	if kepsekID <= 0 {
// 		return nil, ErrActivityCommentForbidden
// 	}

// 	if req.DailyActivityID <= 0 {
// 		return nil, ErrDailyActivityNotFound
// 	}

// 	commentText := strings.TrimSpace(req.Comment)
// 	if commentText == "" {
// 		return nil, ErrActivityCommentRequired
// 	}

// 	if req.Status != 1 && req.Status != 2 {
// 		return nil, ErrInvalidActivityCommentStatus
// 	}

// 	// Validate Admin
// 	admin, err := s.userRepo.FindByIDWithRole(kepsekID)
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, ErrActivityCommentForbidden
// 		}
// 		return nil, err
// 	}

// 	if admin.Role == nil ||
// 		!strings.EqualFold(strings.TrimSpace(*admin.Role), "Admin") {
// 		return nil, ErrActivityCommentForbidden
// 	}

// 	// Find Daily Activity
// 	activity, err := s.dailyActivityRepo.FindByID(
// 		ctx,
// 		req.DailyActivityID,
// 	)
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, ErrDailyActivityNotFound
// 		}
// 		return nil, err
// 	}

// 	// Find activity owner
// 	owner, err := s.userRepo.FindByID(activity.UserID)
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, ErrCommentOwnerNotFound
// 		}
// 		return nil, err
// 	}

// 	now := time.Now()

// 	comment := &entity.ActivityComment{
// 		DailyActivityID: req.DailyActivityID,
// 		KepsekID:        kepsekID,
// 		Comment:         commentText,
// 		Status:          req.Status,
// 		WANotified:      false,
// 		CreatedBy:       &kepsekID,
// 		CreatedAt:       now,
// 		UpdatedBy:       &kepsekID,
// 		UpdatedAt:       now,
// 		IsDeleted:       false,
// 	}

// 	// Save comment first.
// 	// WhatsApp failure must not cancel comment creation.
// 	if err := s.commentRepo.Create(ctx, comment); err != nil {
// 		return nil, err
// 	}

// 	// WhatsApp notification disabled
// 	if !s.cfg.ActivityCommentWhatsAppEnabled {
// 		return comment, nil
// 	}

// 	// Skip notification for inactive users or empty phone numbers.
// 	if !owner.IsActive || strings.TrimSpace(owner.PhoneNumber) == "" {
// 		return comment, nil
// 	}

// 	message := fmt.Sprintf(
// 		"Halo %s,\n\nKepala Sekolah memberikan komentar pada aktivitas harian Anda tanggal %s:\n\n%s\n\nSilakan buka aplikasi Presensi untuk melihat detail aktivitas.",
// 		owner.FullName,
// 		activity.ActivityDate.Format("02-01-2006"),
// 		commentText,
// 	)

// 	// Send WhatsApp notification
// 	if err := s.notificationSvc.SendWhatsAppToUser(
// 		ctx,
// 		owner.ID,
// 		owner.PhoneNumber,
// 		message,
// 	); err != nil {
// 		log.Printf(
// 			"WARNING: failed to send activity comment WhatsApp. comment_id=%d user_id=%d error=%v",
// 			comment.ID,
// 			owner.ID,
// 			err,
// 		)

// 		return comment, nil
// 	}

// 	// Update notification status after successful delivery
// 	if err := s.commentRepo.UpdateWANotified(
// 		ctx,
// 		comment.ID,
// 		true,
// 	); err != nil {
// 		log.Printf(
// 			"WARNING: WhatsApp sent but failed to update wa_notified. comment_id=%d error=%v",
// 			comment.ID,
// 			err,
// 		)

// 		return comment, nil
// 	}

// 	comment.WANotified = true

// 	return comment, nil
// }

func (s *activityCommentService) GetByDailyActivityID(
	ctx context.Context,
	userID int64,
	dailyActivityID int64,
) ([]*entity.ActivityComment, error) {

	if userID <= 0 {
		return nil, ErrActivityCommentForbidden
	}

	if dailyActivityID <= 0 {
		return nil, ErrDailyActivityNotFound
	}

	user, err := s.userRepo.FindByIDWithRole(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrActivityCommentForbidden
		}
		return nil, err
	}

	activity, err := s.dailyActivityRepo.FindByID(
		ctx,
		dailyActivityID,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDailyActivityNotFound
		}
		return nil, err
	}

	isAdmin := user.Role != nil &&
		strings.EqualFold(strings.TrimSpace(*user.Role), "Admin")

	if !isAdmin {
		isEmployee := user.Role != nil &&
			(strings.EqualFold(strings.TrimSpace(*user.Role), "Guru") ||
				strings.EqualFold(strings.TrimSpace(*user.Role), "Karyawan"))

		if !isEmployee || activity.UserID != userID {
			return nil, ErrActivityCommentForbidden
		}
	}

	return s.commentRepo.FindByDailyActivityID(
		ctx,
		dailyActivityID,
	)
}
