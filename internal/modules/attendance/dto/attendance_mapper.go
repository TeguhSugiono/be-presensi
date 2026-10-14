package dto

import (
	"time"

	"be-presensi/internal/modules/attendance/entity"
)

const (
	dateFormat     = "2006-01-02"
	dateTimeFormat = "2006-01-02 15:04:05"
	scheduleFormat = "15:04:05"
)

func ToAttendanceResponse(
	attendance *entity.Attendance,
) *AttendanceResponse {
	if attendance == nil {
		return nil
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}

	response := &AttendanceResponse{
		ID:             attendance.ID,
		UserID:         attendance.UserID,
		AttendanceDate: attendance.AttendanceDate.In(loc).Format(dateFormat),
		CheckInTime:    attendance.CheckInTime.In(loc).Format(dateTimeFormat),
		CheckInPhoto:   attendance.CheckInPhoto,
		CheckInLat:     attendance.CheckInLat,
		CheckInLng:     attendance.CheckInLng,
		IsLate:         attendance.IsLate,
		CheckOutPhoto:  attendance.CheckOutPhoto,
		CheckOutLat:    attendance.CheckOutLat,
		CheckOutLng:    attendance.CheckOutLng,
		IsOutOfArea:    attendance.IsOutOfArea,
		OutOfAreaNote:  attendance.OutOfAreaNote,
		CreatedAt:      attendance.CreatedAt.In(loc).Format(dateTimeFormat),
	}

	if attendance.ScheduledStartTime != nil {
		value := formatScheduleTime(*attendance.ScheduledStartTime)
		response.ScheduledStartTime = &value
	}

	if attendance.ScheduledEndTime != nil {
		value := formatScheduleTime(*attendance.ScheduledEndTime)
		response.ScheduledEndTime = &value
	}

	if attendance.CheckOutTime != nil {
		value := attendance.CheckOutTime.In(loc).Format(dateTimeFormat)
		response.CheckOutTime = &value
	}

	if attendance.UpdatedAt != nil {
		value := attendance.UpdatedAt.In(loc).Format(dateTimeFormat)
		response.UpdatedAt = &value
	}

	return response
}

func formatScheduleTime(value string) string {
	parsed, err := time.Parse("15:04:05", value)
	if err == nil {
		return parsed.Format(scheduleFormat)
	}

	parsed, err = time.Parse("15:04", value)
	if err == nil {
		return parsed.Format(scheduleFormat)
	}

	return value
}
