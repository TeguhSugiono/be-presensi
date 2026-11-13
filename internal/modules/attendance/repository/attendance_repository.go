package repository

import (
	"context"
	"strings"
	"time"

	"be-presensi/internal/modules/attendance/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AttendanceListItem struct {
	ID                 *int64     `gorm:"column:id"`
	UserID             int64      `gorm:"column:user_id"`
	UserName           string     `gorm:"column:user_name"`
	AttendanceDate     *time.Time `gorm:"column:attendance_date"`
	ScheduledStartTime *string    `gorm:"column:scheduled_start_time"`
	ScheduledEndTime   *string    `gorm:"column:scheduled_end_time"`
	CheckInTime        *time.Time `gorm:"column:check_in_time"`
	CheckInPhoto       *string    `gorm:"column:check_in_photo"`
	CheckInLat         *float64   `gorm:"column:check_in_lat"`
	CheckInLng         *float64   `gorm:"column:check_in_lng"`
	KeteranganIn       *string    `gorm:"column:keterangan_in"`
	IsLate             *bool      `gorm:"column:is_late"`
	CheckOutTime       *time.Time `gorm:"column:check_out_time"`
	CheckOutPhoto      *string    `gorm:"column:check_out_photo"`
	CheckOutLat        *float64   `gorm:"column:check_out_lat"`
	CheckOutLng        *float64   `gorm:"column:check_out_lng"`
	KeteranganOut      *string    `gorm:"column:keterangan_out"`
	IsOutOfArea        *bool      `gorm:"column:is_out_of_area"`
	OutOfAreaNote      *string    `gorm:"column:out_of_area_note"`
	CreatedAt          *time.Time `gorm:"column:created_at"`
	UpdatedAt          *time.Time `gorm:"column:updated_at"`
}

type AttendanceSummary struct {
	TotalUsers  int64 `json:"total_users" gorm:"column:total_users"`
	ClockIn     int64 `json:"clock_in" gorm:"column:clock_in"`
	ClockOut    int64 `json:"clock_out" gorm:"column:clock_out"`
	NotAttended int64 `json:"not_attended" gorm:"column:not_attended"`
}

type AttendanceRepository interface {
	WithTx(tx *gorm.DB) AttendanceRepository

	FindAll(ctx context.Context) ([]entity.Attendance, error)
	FindByID(ctx context.Context, id int64) (*entity.Attendance, error)
	FindByUserAndDate(ctx context.Context, userID int64, attendanceDate time.Time) (*entity.Attendance, error)
	Create(ctx context.Context, attendance *entity.Attendance) error
	Update(ctx context.Context, attendance *entity.Attendance) error
	Delete(ctx context.Context, attendance *entity.Attendance) error
	// FindAllPaginated(
	// 	ctx context.Context,
	// 	userID int64,
	// 	isAdmin bool,
	// 	search string,
	// 	limit, offset int,
	// ) ([]*entity.Attendance, int64, error)

	// FindAllPaginated(
	// 	ctx context.Context,
	// 	userID int64,
	// 	isAdmin bool,
	// 	search string,
	// 	date *time.Time,
	// 	limit, offset int,
	// ) ([]*AttendanceListItem, int64, error)

	FindAllPaginated(
		ctx context.Context,
		userID int64,
		isAdmin bool,
		search string,
		dateStart *time.Time,
		dateEnd *time.Time,
		limit, offset int,
	) ([]*AttendanceListItem, int64, error)

	FindByIDForUpdate(
		ctx context.Context,
		id int64,
	) (*entity.Attendance, error)

	FindByUserAndDateForUpdate(
		ctx context.Context,
		userID int64,
		attendanceDate time.Time,
	) (*entity.Attendance, error)

	GetSummary(
		ctx context.Context,
		date time.Time,
	) (*AttendanceSummary, error)
}

type attendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db: db}
}

func (r *attendanceRepository) WithTx(tx *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db: tx}
}

func (r *attendanceRepository) FindAll(ctx context.Context) ([]entity.Attendance, error) {
	var attendances []entity.Attendance

	err := r.db.WithContext(ctx).
		Where("is_deleted = ?", false).
		Order("attendance_date DESC, id DESC").
		Find(&attendances).Error

	return attendances, err
}

func (r *attendanceRepository) FindByID(
	ctx context.Context,
	id int64,
) (*entity.Attendance, error) {
	var attendance entity.Attendance

	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&attendance).Error
	if err != nil {
		return nil, err
	}

	return &attendance, nil
}

func (r *attendanceRepository) FindByUserAndDate(
	ctx context.Context,
	userID int64,
	attendanceDate time.Time,
) (*entity.Attendance, error) {
	var attendance entity.Attendance

	err := r.db.WithContext(ctx).
		Where(
			"user_id = ? AND attendance_date = ? AND is_deleted = ?",
			userID,
			attendanceDate.Format("2006-01-02"),
			false,
		).
		First(&attendance).Error
	if err != nil {
		return nil, err
	}

	return &attendance, nil
}

func (r *attendanceRepository) Create(
	ctx context.Context,
	attendance *entity.Attendance,
) error {
	return r.db.WithContext(ctx).Create(attendance).Error
}

func (r *attendanceRepository) Update(
	ctx context.Context,
	attendance *entity.Attendance,
) error {
	return r.db.WithContext(ctx).Save(attendance).Error
}

func (r *attendanceRepository) Delete(
	ctx context.Context,
	attendance *entity.Attendance,
) error {
	return r.db.WithContext(ctx).
		Model(&entity.Attendance{}).
		Where("id = ? AND is_deleted = ?", attendance.ID, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"updated_by": attendance.UpdatedBy,
			"updated_at": attendance.UpdatedAt,
		}).Error
}

func (r *attendanceRepository) FindAllPaginated(
	ctx context.Context,
	userID int64,
	isAdmin bool,
	search string,
	dateStart *time.Time,
	dateEnd *time.Time,
	limit, offset int,
) ([]*AttendanceListItem, int64, error) {
	var items []*AttendanceListItem
	var total int64

	var query *gorm.DB

	// Jika rentang tanggal dipilih, tampilkan semua pegawai aktif
	// pada setiap tanggal, termasuk yang belum melakukan absensi.
	if dateStart != nil && dateEnd != nil {
		unionSQL := `
			SELECT
				attendances.id,
				attendances.user_id,
				users.full_name AS user_name,
				attendances.attendance_date,
				attendances.scheduled_start_time,
				attendances.scheduled_end_time,
				attendances.check_in_time,
				attendances.check_in_photo,
				attendances.check_in_lat,
				attendances.check_in_lng,
				attendances.keterangan_in,
				attendances.is_late,
				attendances.check_out_time,
				attendances.check_out_photo,
				attendances.check_out_lat,
				attendances.check_out_lng,
				attendances.keterangan_out,
				attendances.is_out_of_area,
				attendances.out_of_area_note,
				attendances.created_at,
				attendances.updated_at
			FROM attendances
			INNER JOIN users ON users.id = attendances.user_id
			WHERE attendances.is_deleted = false
			  AND users.is_deleted = false
			  AND attendances.attendance_date BETWEEN ?::date AND ?::date
			  AND (? = TRUE OR attendances.user_id = ?)

			UNION ALL

			SELECT
				NULL::bigint AS id,
				users.id AS user_id,
				users.full_name AS user_name,
				calendar.attendance_date::date AS attendance_date,
				NULL::time AS scheduled_start_time,
				NULL::time AS scheduled_end_time,
				NULL::timestamp AS check_in_time,
				NULL::text AS check_in_photo,
				NULL::numeric AS check_in_lat,
				NULL::numeric AS check_in_lng,
				NULL::text AS keterangan_in,
				NULL::boolean AS is_late,
				NULL::timestamp AS check_out_time,
				NULL::text AS check_out_photo,
				NULL::numeric AS check_out_lat,
				NULL::numeric AS check_out_lng,
				NULL::text AS keterangan_out,
				NULL::boolean AS is_out_of_area,
				NULL::text AS out_of_area_note,
				NULL::timestamp AS created_at,
				NULL::timestamp AS updated_at
			FROM generate_series(
				?::date,
				?::date,
				INTERVAL '1 day'
			) AS calendar(attendance_date)
			CROSS JOIN users
			INNER JOIN roles ON roles.id = users.role_id
			WHERE users.is_active = true
			  AND users.is_deleted = false
			  AND roles.is_deleted = false
			  AND LOWER(roles.name) IN ('guru', 'karyawan')
			  AND (? = TRUE OR users.id = ?)
			  AND NOT EXISTS (
				  SELECT 1
				  FROM attendances
				  WHERE attendances.user_id = users.id
					AND attendances.attendance_date = calendar.attendance_date::date
					AND attendances.is_deleted = false
			  )
		`

		query = r.db.WithContext(ctx).
			Table("(?) AS attendance_items",
				r.db.Raw(
					unionSQL,
					dateStart, dateEnd, isAdmin, userID,
					dateStart, dateEnd, isAdmin, userID,
				),
			)
	} else {
		// Tanpa filter tanggal, perilaku lama tetap dipertahankan:
		// hanya menampilkan record absensi yang benar-benar ada.
		query = r.db.WithContext(ctx).
			Model(&entity.Attendance{}).
			Joins("LEFT JOIN users ON users.id = attendances.user_id").
			Where("attendances.is_deleted = ?", false).
			Where("users.is_deleted = ?", false)

		if !isAdmin {
			query = query.Where("attendances.user_id = ?", userID)
		}

		query = query.Select(`
			attendances.id,
			attendances.user_id,
			users.full_name AS user_name,
			attendances.attendance_date,
			attendances.scheduled_start_time,
			attendances.scheduled_end_time,
			attendances.check_in_time,
			attendances.check_in_photo,
			attendances.check_in_lat,
			attendances.check_in_lng,
			attendances.keterangan_in,
			attendances.is_late,
			attendances.check_out_time,
			attendances.check_out_photo,
			attendances.check_out_lat,
			attendances.check_out_lng,
			attendances.keterangan_out,
			attendances.is_out_of_area,
			attendances.out_of_area_note,
			attendances.created_at,
			attendances.updated_at
		`)
	}

	// Search diterapkan pada hasil gabungan.
	// if search != "" {
	// 	keyword := "%" + strings.TrimSpace(search) + "%"

	// 	query = query.Where(
	// 		`(
	// 			user_name ILIKE ?
	// 			OR CAST(user_id AS TEXT) ILIKE ?
	// 			OR CAST(attendance_date AS TEXT) ILIKE ?
	// 		)`,
	// 		keyword,
	// 		keyword,
	// 		keyword,
	// 	)
	// }

	// Search diterapkan pada hasil gabungan.
	if search != "" {
		keyword := "%" + strings.TrimSpace(search) + "%"

		query = query.Where(
			`(
            user_name ILIKE ?
            OR CAST(user_id AS TEXT) ILIKE ?
            OR CAST(attendance_date AS TEXT) ILIKE ?
            OR keterangan_in ILIKE ?
            OR keterangan_out ILIKE ?
            OR out_of_area_note ILIKE ?
        )`,
			keyword,
			keyword,
			keyword,
			keyword,
			keyword,
			keyword,
		)
	}

	// Hitung seluruh hasil sebelum pagination.
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("attendance_date DESC").
		Order("id DESC NULLS LAST").
		Order("user_name ASC").
		Limit(limit).
		Offset(offset).
		Scan(&items).Error

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// func (r *attendanceRepository) FindAllPaginated(
// 	ctx context.Context,
// 	userID int64,
// 	isAdmin bool,
// 	search string,
// 	date *time.Time,
// 	limit, offset int,
// ) ([]*AttendanceListItem, int64, error) {
// 	var items []*AttendanceListItem
// 	var total int64

// 	var query *gorm.DB

// 	// Jika Admin memilih tanggal tertentu,
// 	// gabungkan attendance yang sudah ada dengan user yang belum absen.
// 	if date != nil {
// 		unionSQL := `
//             SELECT
//                 attendances.id,
//                 attendances.user_id,
//                 users.full_name AS user_name,
//                 attendances.attendance_date,
//                 attendances.scheduled_start_time,
//                 attendances.scheduled_end_time,
//                 attendances.check_in_time,
//                 attendances.check_in_photo,
//                 attendances.check_in_lat,
//                 attendances.check_in_lng,
//                 attendances.keterangan_in,
//                 attendances.is_late,
//                 attendances.check_out_time,
//                 attendances.check_out_photo,
//                 attendances.check_out_lat,
//                 attendances.check_out_lng,
//                 attendances.keterangan_out,
//                 attendances.is_out_of_area,
//                 attendances.out_of_area_note,
//                 attendances.created_at,
//                 attendances.updated_at
//             FROM attendances
//             INNER JOIN users ON users.id = attendances.user_id
//             WHERE attendances.is_deleted = false
//               AND users.is_deleted = false
//               AND attendances.attendance_date = ?::date AND (? = TRUE OR attendances.user_id = ?)

//             UNION ALL

//             SELECT
//                 NULL::bigint AS id,
//                 users.id AS user_id,
//                 users.full_name AS user_name,
//                 ?::date AS attendance_date,
//                 NULL::time AS scheduled_start_time,
//                 NULL::time AS scheduled_end_time,
//                 NULL::timestamp AS check_in_time,
//                 NULL::text AS check_in_photo,
//                 NULL::numeric AS check_in_lat,
//                 NULL::numeric AS check_in_lng,
//                 NULL::text AS keterangan_in,
//                 NULL::boolean AS is_late,
//                 NULL::timestamp AS check_out_time,
//                 NULL::text AS check_out_photo,
//                 NULL::numeric AS check_out_lat,
//                 NULL::numeric AS check_out_lng,
//                 NULL::text AS keterangan_out,
//                 NULL::boolean AS is_out_of_area,
//                 NULL::text AS out_of_area_note,
//                 NULL::timestamp AS created_at,
//                 NULL::timestamp AS updated_at
//             FROM users
//             INNER JOIN roles ON roles.id = users.role_id
//             WHERE users.is_active = true
//               AND users.is_deleted = false
//               AND roles.is_deleted = false
//               AND LOWER(roles.name) IN ('guru', 'karyawan')
// 			  AND (? = TRUE OR users.id = ?)
//               AND NOT EXISTS (
//                   SELECT 1
//                   FROM attendances
//                   WHERE attendances.user_id = users.id
//                     AND attendances.attendance_date = ?::date
//                     AND attendances.is_deleted = false
//               )
//         `

// 		query = r.db.WithContext(ctx).
// 			Table("(?) AS attendance_items",
// 				r.db.Raw(
// 					unionSQL,
// 					date, isAdmin, userID,
// 					date, isAdmin, userID,
// 					date,
// 				),
// 			)
// 	} else {
// 		// Perilaku lama tetap dipertahankan jika tanggal tidak dipilih
// 		// atau pengguna bukan Admin.
// 		query = r.db.WithContext(ctx).
// 			Model(&entity.Attendance{}).
// 			Joins("LEFT JOIN users ON users.id = attendances.user_id").
// 			Where("attendances.is_deleted = ?", false).
// 			Where("users.is_deleted = ?", false)

// 		if !isAdmin {
// 			query = query.Where("attendances.user_id = ?", userID)
// 		}

// 		if date != nil {
// 			query = query.Where(
// 				"attendances.attendance_date = ?",
// 				date.Format("2006-01-02"),
// 			)
// 		}

// 		query = query.Select(`
//             attendances.id,
//             attendances.user_id,
//             users.full_name AS user_name,
//             attendances.attendance_date,
//             attendances.scheduled_start_time,
//             attendances.scheduled_end_time,
//             attendances.check_in_time,
//             attendances.check_in_photo,
//             attendances.check_in_lat,
//             attendances.check_in_lng,
//             attendances.keterangan_in,
//             attendances.is_late,
//             attendances.check_out_time,
//             attendances.check_out_photo,
//             attendances.check_out_lat,
//             attendances.check_out_lng,
//             attendances.keterangan_out,
//             attendances.is_out_of_area,
//             attendances.out_of_area_note,
//             attendances.created_at,
//             attendances.updated_at
//         `)
// 	}

// 	if search != "" {
// 		keyword := "%" + strings.TrimSpace(search) + "%"

// 		query = query.Where(
// 			`(
//                 user_name ILIKE ?
//                 OR CAST(user_id AS TEXT) ILIKE ?
//                 OR CAST(attendance_date AS TEXT) ILIKE ?
//             )`,
// 			keyword,
// 			keyword,
// 			keyword,
// 		)
// 	}

// 	// Hitung seluruh hasil sebelum pagination.
// 	if err := query.Count(&total).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	err := query.
// 		Order("attendance_date DESC").
// 		Order("id DESC NULLS LAST").
// 		Order("user_name ASC").
// 		Limit(limit).
// 		Offset(offset).
// 		Scan(&items).Error

// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	return items, total, nil
// }

// func (r *attendanceRepository) FindAllPaginated(
// 	ctx context.Context,
// 	userID int64,
// 	isAdmin bool,
// 	search string,
// 	limit, offset int,
// ) ([]*entity.Attendance, int64, error) {
// 	var attendances []*entity.Attendance
// 	var total int64

// 	query := r.db.WithContext(ctx).
// 		Model(&entity.Attendance{}).
// 		Joins("LEFT JOIN users ON users.id = attendances.user_id").
// 		Where("attendances.is_deleted = ?", false).
// 		Where("users.is_deleted = ?", false)

// 	// Admin melihat seluruh data.
// 	// User biasa hanya melihat data absensinya sendiri.
// 	if !isAdmin {
// 		query = query.Where("attendances.user_id = ?", userID)
// 	}

// 	if search != "" {
// 		keyword := "%" + strings.TrimSpace(search) + "%"

// 		query = query.Where(
// 			`(
//                 users.full_name ILIKE ?
//                 OR CAST(attendances.user_id AS TEXT) ILIKE ?
//                 OR CAST(attendances.attendance_date AS TEXT) ILIKE ?
//             )`,
// 			keyword,
// 			keyword,
// 			keyword,
// 		)
// 	}

// 	if err := query.Count(&total).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	err := query.
// 		Select("attendances.*, users.full_name AS user_name").
// 		Order("attendances.attendance_date DESC, attendances.id DESC").
// 		Limit(limit).
// 		Offset(offset).
// 		Find(&attendances).Error
// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	return attendances, total, nil
// }

func (r *attendanceRepository) FindByIDForUpdate(
	ctx context.Context,
	id int64,
) (*entity.Attendance, error) {
	var attendance entity.Attendance

	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&attendance).Error
	if err != nil {
		return nil, err
	}

	return &attendance, nil
}

func (r *attendanceRepository) FindByUserAndDateForUpdate(
	ctx context.Context,
	userID int64,
	attendanceDate time.Time,
) (*entity.Attendance, error) {
	var attendance entity.Attendance

	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			"user_id = ? AND attendance_date = ? AND is_deleted = ?",
			userID,
			attendanceDate.Format("2006-01-02"),
			false,
		).
		First(&attendance).Error
	if err != nil {
		return nil, err
	}

	return &attendance, nil
}

func (r *attendanceRepository) GetSummary(
	ctx context.Context,
	date time.Time,
) (*AttendanceSummary, error) {
	var summary AttendanceSummary

	query := `
        SELECT
            COUNT(DISTINCT users.id) AS total_users,

            COUNT(DISTINCT users.id) FILTER (
                WHERE attendances.check_in_time IS NOT NULL
            ) AS clock_in,

            COUNT(DISTINCT users.id) FILTER (
                WHERE attendances.check_out_time IS NOT NULL
            ) AS clock_out,

            COUNT(DISTINCT users.id) FILTER (
                WHERE attendances.check_in_time IS NULL
            ) AS not_attended

        FROM users
        INNER JOIN roles
            ON roles.id = users.role_id

        LEFT JOIN attendances
            ON attendances.user_id = users.id
            AND attendances.attendance_date = ?::date
            AND attendances.is_deleted = false

        WHERE users.is_active = true
            AND users.is_deleted = false
            AND roles.is_deleted = false
            AND LOWER(roles.name) IN ('guru', 'karyawan')
    `

	err := r.db.WithContext(ctx).
		Raw(query, date.Format("2006-01-02")).
		Scan(&summary).Error

	if err != nil {
		return nil, err
	}

	return &summary, nil
}
