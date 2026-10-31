package dto

import (
	"strconv"

	attendanceEntity "be-presensi/internal/modules/attendance/entity"
)

func ToAttendanceTodayResponse(
	attendance *attendanceEntity.Attendance,
	appBaseURL string,
) *AttendanceTodayResponse {
	if attendance == nil {
		return nil
	}

	attendanceID := attendance.ID
	attendanceDate := attendance.AttendanceDate.Format("2006-01-02")
	checkInTime := attendance.CheckInTime.Format("2006-01-02 15:04:05")

	checkInLat := attendance.CheckInLat
	checkInLng := attendance.CheckInLng

	photoURLIn := appBaseURL +
		"/api/attendances/" +
		strconv.FormatInt(attendance.ID, 10) +
		"/check-in-photo"

	response := &AttendanceTodayResponse{
		ID:             &attendanceID,
		AttendanceDate: &attendanceDate,
		CheckInTime:    &checkInTime,
		PhotoURLIn:     &photoURLIn,
		LatitudeIn:     &checkInLat,
		LongitudeIn:    &checkInLng,
		KeteranganIn:   attendance.KeteranganIn,
	}

	if attendance.CheckOutTime != nil {
		checkOutTime := attendance.CheckOutTime.Format("2006-01-02 15:04:05")

		photoURLOut := appBaseURL +
			"/api/attendances/" +
			strconv.FormatInt(attendance.ID, 10) +
			"/check-out-photo"

		response.CheckOutTime = &checkOutTime
		response.PhotoURLOut = &photoURLOut
		response.LatitudeOut = attendance.CheckOutLat
		response.LongitudeOut = attendance.CheckOutLng
		response.KeteranganOut = attendance.KeteranganOut
	}

	return response
}
