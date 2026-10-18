package entity

import "time"

type Attendance struct {
	ID                 int64      `gorm:"column:id;primaryKey"`
	UserID             int64      `gorm:"column:user_id;not null"`
	AttendanceDate     time.Time  `gorm:"column:attendance_date;type:date;not null"`
	ScheduledStartTime *string    `gorm:"column:scheduled_start_time;type:time"`
	ScheduledEndTime   *string    `gorm:"column:scheduled_end_time;type:time"`
	CheckInTime        time.Time  `gorm:"column:check_in_time;not null"`
	CheckInPhoto       string     `gorm:"column:check_in_photo;type:text;not null"`
	CheckInLat         float64    `gorm:"column:check_in_lat;type:numeric;not null"`
	CheckInLng         float64    `gorm:"column:check_in_lng;type:numeric;not null"`
	IsLate             bool       `gorm:"column:is_late;not null;default:false"`
	CheckOutTime       *time.Time `gorm:"column:check_out_time"`
	CheckOutPhoto      *string    `gorm:"column:check_out_photo;type:text"`
	CheckOutLat        *float64   `gorm:"column:check_out_lat;type:numeric"`
	CheckOutLng        *float64   `gorm:"column:check_out_lng;type:numeric"`
	IsOutOfArea        bool       `gorm:"column:is_out_of_area;not null;default:false"`
	OutOfAreaNote      *string    `gorm:"column:out_of_area_note;type:text"`
	CreatedBy          *int64     `gorm:"column:created_by"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null"`
	UpdatedBy          *int64     `gorm:"column:updated_by"`
	UpdatedAt          *time.Time `gorm:"column:updated_at;autoUpdateTime:false"`
	IsDeleted          bool       `gorm:"column:is_deleted;not null;default:false"`
	UserName           string     `gorm:"column:user_name;->" json:"-"`
}

func (Attendance) TableName() string {
	return "attendances"
}
