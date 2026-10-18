package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"be-presensi/internal/modules/attendance/dto"
	attendanceEntity "be-presensi/internal/modules/attendance/entity"
	attendanceRepository "be-presensi/internal/modules/attendance/repository"
	dailyActivityRepository "be-presensi/internal/modules/daily_activity/repository"
	roleRepository "be-presensi/internal/modules/role/repository"
	schoolLocationRepository "be-presensi/internal/modules/school_location/repository"
	userRepository "be-presensi/internal/modules/user/repository"
	workScheduleRepository "be-presensi/internal/modules/work_schedule/repository"

	"gorm.io/gorm"
)

var (
	ErrAttendanceUserNotFound      = errors.New("user not found")
	ErrAttendanceRoleNotFound      = errors.New("user role not found")
	ErrAttendanceScheduleRequired  = errors.New("work schedule is required")
	ErrAttendanceAlreadyExists     = errors.New("attendance already exists for this date")
	ErrAttendanceNotFound          = errors.New("attendance not found")
	ErrAttendanceAlreadyCheckedOut = errors.New("attendance has already been checked out")
	ErrAttendanceScheduleNotEnded  = errors.New("work schedule has not ended")
	ErrDailyActivityRequired       = errors.New("daily activity is required before check-out")
	ErrDailyActivityLocked         = errors.New("daily activity is already locked")
	ErrSchoolLocationUnavailable   = errors.New("no active school location is available")
	ErrOutOfAreaNoteRequired       = errors.New("out-of-area note is required")
	ErrInvalidCoordinates          = errors.New("invalid GPS coordinates")
	ErrDailyActivityNotSubmitted   = errors.New("daily activity must be submitted before check-out")
	ErrAttendancePhotoNotFound     = errors.New("attendance photo not found")
)

type AttendanceService struct {
	db                 *gorm.DB
	attendanceRepo     attendanceRepository.AttendanceRepository
	dailyActivityRepo  dailyActivityRepository.DailyActivityRepository
	userRepo           userRepository.UserRepository
	roleRepo           roleRepository.RoleRepository
	workScheduleRepo   workScheduleRepository.WorkScheduleRepository
	schoolLocationRepo schoolLocationRepository.SchoolLocationRepository
}

func NewAttendanceService(
	db *gorm.DB,
	attendanceRepo attendanceRepository.AttendanceRepository,
	dailyActivityRepo dailyActivityRepository.DailyActivityRepository,
	userRepo userRepository.UserRepository,
	roleRepo roleRepository.RoleRepository,
	workScheduleRepo workScheduleRepository.WorkScheduleRepository,
	schoolLocationRepo schoolLocationRepository.SchoolLocationRepository,
) *AttendanceService {
	return &AttendanceService{
		db:                 db,
		attendanceRepo:     attendanceRepo,
		dailyActivityRepo:  dailyActivityRepo,
		userRepo:           userRepo,
		roleRepo:           roleRepo,
		workScheduleRepo:   workScheduleRepo,
		schoolLocationRepo: schoolLocationRepo,
	}
}

func (s *AttendanceService) FindAllPaginated(
	ctx context.Context,
	search string,
	page int,
	perPage int,
) ([]*attendanceEntity.Attendance, int64, error) {
	if page < 1 {
		page = 1
	}

	if perPage < 1 {
		perPage = 10
	}

	if perPage > 100 {
		perPage = 100
	}

	offset := (page - 1) * perPage

	return s.attendanceRepo.FindAllPaginated(
		ctx,
		strings.TrimSpace(search),
		perPage,
		offset,
	)
}

// GetAttendancePhoto returns a private attendance photo path.
// Only the attendance owner or an Admin can access it.
func (s *AttendanceService) GetAttendancePhoto(
	ctx context.Context,
	userID int64,
	attendanceID int64,
	photoType string,
) (string, error) {
	if userID <= 0 || attendanceID <= 0 {
		return "", ErrAttendanceNotFound
	}

	if photoType != "check-in" && photoType != "check-out" {
		return "", ErrAttendancePhotoNotFound
	}

	// Validate the authenticated user.
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrAttendanceUserNotFound
		}
		return "", err
	}

	if user == nil || !user.IsActive || user.IsDeleted {
		return "", ErrAttendanceUserNotFound
	}

	// Load attendance.
	attendance, err := s.attendanceRepo.FindByID(ctx, attendanceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrAttendanceNotFound
		}
		return "", err
	}

	if attendance == nil || attendance.IsDeleted {
		return "", ErrAttendanceNotFound
	}

	// Determine whether the authenticated user is an Admin.
	isAdmin := false

	role, err := s.roleRepo.FindByID(ctx, user.RoleID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}
	} else if role != nil && !role.IsDeleted &&
		strings.EqualFold(strings.TrimSpace(role.Name), "Admin") {
		isAdmin = true
	}

	// Hide attendance existence from unauthorized users.
	if attendance.UserID != userID && !isAdmin {
		return "", ErrAttendanceNotFound
	}

	// Return only the requested photo path.
	var photoPath string

	switch photoType {
	case "check-in":
		photoPath = attendance.CheckInPhoto
	case "check-out":
		if attendance.CheckOutPhoto != nil {
			photoPath = *attendance.CheckOutPhoto
		}
	}

	if strings.TrimSpace(photoPath) == "" {
		return "", ErrAttendancePhotoNotFound
	}

	return photoPath, nil
}

// CheckIn records attendance using the authenticated user's ID.
// userID must come from JWT middleware, not from the request body.
func (s *AttendanceService) CheckIn(
	ctx context.Context,
	userID int64,
	req dto.CheckInAttendanceRequest,
) (*attendanceEntity.Attendance, error) {
	serverTime := time.Now()

	clientTime, err := validateClientAttendanceTime(
		req.CheckInDate,
		req.CheckInTime,
		serverTime,
	)
	if err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}
	clientTime = clientTime.In(loc)
	attendanceDate := time.Date(
		clientTime.Year(),
		clientTime.Month(),
		clientTime.Day(),
		0, 0, 0, 0,
		loc,
	)

	if strings.TrimSpace(req.CheckInPhoto) == "" {
		return nil, errors.New("check-in photo is required")
	}

	if err := validateCoordinates(req.CheckInLat, req.CheckInLng); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAttendanceUserNotFound
		}
		return nil, err
	}
	if user == nil || !user.IsActive || user.IsDeleted {
		return nil, ErrAttendanceUserNotFound
	}

	role, err := s.roleRepo.FindByID(ctx, user.RoleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAttendanceRoleNotFound
		}
		return nil, err
	}
	if role == nil || role.IsDeleted {
		return nil, ErrAttendanceRoleNotFound
	}

	var scheduleStart *string
	var scheduleEnd *string
	isLate := false

	if !role.IsTeacher {
		if user.WorkScheduleID == nil {
			return nil, ErrAttendanceScheduleRequired
		}

		schedule, err := s.workScheduleRepo.FindByID(ctx, *user.WorkScheduleID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrAttendanceScheduleRequired
			}
			return nil, err
		}
		if schedule == nil || schedule.IsDeleted || schedule.IsFlexible {
			return nil, ErrAttendanceScheduleRequired
		}
		if schedule.StartTime == nil || schedule.EndTime == nil {
			return nil, ErrAttendanceScheduleRequired
		}

		startTime, err := parseScheduleTime(*schedule.StartTime)
		if err != nil {
			return nil, err
		}

		scheduleStart = schedule.StartTime
		scheduleEnd = schedule.EndTime

		checkInClock := clientTime.Hour()*3600 +
			clientTime.Minute()*60 +
			clientTime.Second()

		isLate = checkInClock > startTime
	}

	existing, err := s.attendanceRepo.FindByUserAndDate(
		ctx,
		userID,
		attendanceDate,
	)
	if err == nil && existing != nil {
		return nil, ErrAttendanceAlreadyExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	createdAt := time.Now()
	createdBy := userID

	attendance := &attendanceEntity.Attendance{
		UserID:             userID,
		AttendanceDate:     attendanceDate,
		ScheduledStartTime: scheduleStart,
		ScheduledEndTime:   scheduleEnd,
		CheckInTime:        clientTime,
		CheckInPhoto:       req.CheckInPhoto,
		CheckInLat:         req.CheckInLat,
		CheckInLng:         req.CheckInLng,
		IsLate:             isLate,
		IsOutOfArea:        false,
		CreatedBy:          &createdBy,
		CreatedAt:          createdAt,
		IsDeleted:          false,
	}

	if err := s.attendanceRepo.Create(ctx, attendance); err != nil {
		// The database's unique (user_id, attendance_date) constraint
		// remains the final protection against concurrent check-ins.
		return nil, err
	}

	return attendance, nil
}

// CheckOut updates attendance and locks the daily activity in one transaction.
func (s *AttendanceService) CheckOut(
	ctx context.Context,
	userID int64,
	attendanceID int64,
	req dto.CheckOutAttendanceRequest,
) (*attendanceEntity.Attendance, error) {
	serverTime := time.Now()

	clientTime, err := validateClientAttendanceTime(
		req.CheckOutDate,
		req.CheckOutTime,
		serverTime,
	)
	if err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}
	clientTime = clientTime.In(loc)

	if strings.TrimSpace(req.CheckOutPhoto) == "" {
		return nil, errors.New("check-out photo is required")
	}
	if req.CheckOutLat == nil || req.CheckOutLng == nil {
		return nil, ErrInvalidCoordinates
	}
	if err := validateCoordinates(*req.CheckOutLat, *req.CheckOutLng); err != nil {
		return nil, err
	}

	// Resolve the user's role before entering the transaction.
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAttendanceUserNotFound
		}
		return nil, err
	}
	if user == nil || !user.IsActive || user.IsDeleted {
		return nil, ErrAttendanceUserNotFound
	}

	role, err := s.roleRepo.FindByID(ctx, user.RoleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAttendanceRoleNotFound
		}
		return nil, err
	}
	if role == nil || role.IsDeleted {
		return nil, ErrAttendanceRoleNotFound
	}

	var updatedAttendance *attendanceEntity.Attendance

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txAttendanceRepo := s.attendanceRepo.WithTx(tx)
		txDailyActivityRepo := s.dailyActivityRepo.WithTx(tx)

		attendance, err := txAttendanceRepo.FindByIDForUpdate(
			ctx,
			attendanceID,
		)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrAttendanceNotFound
			}
			return err
		}

		// Never allow one user to check out another user's attendance.
		if attendance.UserID != userID {
			return ErrAttendanceNotFound
		}
		if attendance.CheckOutTime != nil {
			return ErrAttendanceAlreadyCheckedOut
		}

		if !role.IsTeacher {
			if attendance.ScheduledEndTime == nil {
				return ErrAttendanceScheduleRequired
			}

			endTime, err := parseScheduleTime(*attendance.ScheduledEndTime)
			if err != nil {
				return err
			}

			checkOutClock := clientTime.Hour()*3600 +
				clientTime.Minute()*60 +
				clientTime.Second()

			if checkOutClock < endTime {
				return ErrAttendanceScheduleNotEnded
			}
		}

		activity, err := txDailyActivityRepo.FindByAttendanceIDForUpdate(
			ctx,
			attendance.ID,
			userID,
		)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrDailyActivityRequired
			}
			return err
		}
		if activity == nil {
			return ErrDailyActivityRequired
		}
		if strings.TrimSpace(activity.Description) == "" ||
			(activity.Status != 2 && activity.Status != 3) {
			return ErrDailyActivityNotSubmitted
		}

		outOfArea, err := s.isOutsideActiveSchoolLocations(
			ctx,
			*req.CheckOutLat,
			*req.CheckOutLng,
		)
		if err != nil {
			return err
		}

		var note *string
		if outOfArea {
			if req.OutOfAreaNote == nil ||
				strings.TrimSpace(*req.OutOfAreaNote) == "" {
				return ErrOutOfAreaNoteRequired
			}
			cleanNote := strings.TrimSpace(*req.OutOfAreaNote)
			note = &cleanNote
		}

		checkOutPhoto := req.CheckOutPhoto
		checkOutLat := *req.CheckOutLat
		checkOutLng := *req.CheckOutLng
		updatedBy := userID
		updatedAt := time.Now()

		attendance.CheckOutTime = &clientTime
		attendance.CheckOutPhoto = &checkOutPhoto
		attendance.CheckOutLat = &checkOutLat
		attendance.CheckOutLng = &checkOutLng
		attendance.IsOutOfArea = outOfArea
		attendance.OutOfAreaNote = note
		attendance.UpdatedBy = &updatedBy
		attendance.UpdatedAt = &updatedAt

		if err := txAttendanceRepo.Update(ctx, attendance); err != nil {
			return err
		}

		updatedAttendance = attendance
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Any WhatsApp notification for out-of-area check-outs should be sent
	// after the transaction commits, through a notifier integration.
	return updatedAttendance, nil
}

func parseScheduleTime(value string) (int, error) {
	value = strings.TrimSpace(value)

	for _, layout := range []string{"15:04:05", "15:04"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.Hour()*3600 +
				parsed.Minute()*60 +
				parsed.Second(), nil
		}
	}

	return 0, fmt.Errorf("invalid work schedule time format: %q", value)
}

func validateCoordinates(latitude, longitude float64) error {
	if math.IsNaN(latitude) || math.IsInf(latitude, 0) ||
		math.IsNaN(longitude) || math.IsInf(longitude, 0) ||
		latitude < -90 || latitude > 90 ||
		longitude < -180 || longitude > 180 {
		return ErrInvalidCoordinates
	}
	return nil
}

func (s *AttendanceService) isOutsideActiveSchoolLocations(
	ctx context.Context,
	latitude float64,
	longitude float64,
) (bool, error) {
	locations, err := s.schoolLocationRepo.FindAll(ctx)
	if err != nil {
		return false, err
	}

	hasActiveLocation := false

	for _, location := range locations {
		if location.IsDeleted || !location.IsActive {
			continue
		}

		hasActiveLocation = true

		distance := distanceMeters(
			latitude,
			longitude,
			location.Latitude,
			location.Longitude,
		)

		if distance <= float64(location.RadiusMeters) {
			return false, nil
		}
	}

	if !hasActiveLocation {
		return false, ErrSchoolLocationUnavailable
	}

	return true, nil
}

func distanceMeters(
	lat1, lon1, lat2, lon2 float64,
) float64 {
	const earthRadiusMeters = 6371000.0

	toRadians := func(degrees float64) float64 {
		return degrees * math.Pi / 180
	}

	phi1 := toRadians(lat1)
	phi2 := toRadians(lat2)
	deltaPhi := toRadians(lat2 - lat1)
	deltaLambda := toRadians(lon2 - lon1)

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}
