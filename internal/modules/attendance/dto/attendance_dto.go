package dto

type CheckInAttendanceRequest struct {
	CheckInDate  string  `json:"check_in_date" form:"check_in_date" binding:"required"`
	CheckInTime  string  `json:"check_in_time" form:"check_in_time" binding:"required"`
	CheckInPhoto string  `json:"-" form:"-"`
	CheckInLat   float64 `json:"check_in_lat" form:"check_in_lat"`
	CheckInLng   float64 `json:"check_in_lng" form:"check_in_lng"`
}

type CheckOutAttendanceRequest struct {
	CheckOutDate  string   `json:"check_out_date" form:"check_out_date" binding:"required"`
	CheckOutTime  string   `json:"check_out_time" form:"check_out_time" binding:"required"`
	CheckOutPhoto string   `json:"-" form:"-"`
	CheckOutLat   *float64 `json:"check_out_lat" form:"check_out_lat" binding:"required"`
	CheckOutLng   *float64 `json:"check_out_lng" form:"check_out_lng" binding:"required"`
	OutOfAreaNote *string  `json:"out_of_area_note" form:"out_of_area_note"`
}

type AttendanceResponse struct {
	ID                 int64    `json:"id"`
	UserID             int64    `json:"user_id"`
	AttendanceDate     string   `json:"attendance_date"`
	ScheduledStartTime *string  `json:"scheduled_start_time"`
	ScheduledEndTime   *string  `json:"scheduled_end_time"`
	CheckInTime        string   `json:"check_in_time"`
	CheckInPhoto       string   `json:"check_in_photo"`
	CheckInLat         float64  `json:"check_in_lat"`
	CheckInLng         float64  `json:"check_in_lng"`
	IsLate             bool     `json:"is_late"`
	CheckOutTime       *string  `json:"check_out_time"`
	CheckOutPhoto      *string  `json:"check_out_photo"`
	CheckOutLat        *float64 `json:"check_out_lat"`
	CheckOutLng        *float64 `json:"check_out_lng"`
	IsOutOfArea        bool     `json:"is_out_of_area"`
	OutOfAreaNote      *string  `json:"out_of_area_note"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          *string  `json:"updated_at"`
}
type AttendanceListResponse struct {
	ID                 int64    `json:"id"`
	UserID             int64    `json:"user_id"`
	UserName           string   `json:"user_name"`
	AttendanceDate     string   `json:"attendance_date"`
	ScheduledStartTime *string  `json:"scheduled_start_time"`
	ScheduledEndTime   *string  `json:"scheduled_end_time"`
	CheckInTime        string   `json:"check_in_time"`
	CheckInPhoto       string   `json:"check_in_photo"`
	CheckInLat         float64  `json:"check_in_lat"`
	CheckInLng         float64  `json:"check_in_lng"`
	IsLate             bool     `json:"is_late"`
	CheckOutTime       *string  `json:"check_out_time"`
	CheckOutPhoto      *string  `json:"check_out_photo"`
	CheckOutLat        *float64 `json:"check_out_lat"`
	CheckOutLng        *float64 `json:"check_out_lng"`
	IsOutOfArea        bool     `json:"is_out_of_area"`
	OutOfAreaNote      *string  `json:"out_of_area_note"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          *string  `json:"updated_at"`
}
