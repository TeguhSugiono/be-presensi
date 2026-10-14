package service

import (
	"errors"
	"math"
	"strings"
	"time"
)

const attendanceTimeTolerance = 2 * time.Minute

var (
	ErrInvalidAttendanceTime  = errors.New("unable to process attendance request")
	ErrAttendanceTimeMismatch = errors.New("unable to process attendance request")
)

// Tanggal dan jam dari frontend digabung, lalu dianggap sebagai waktu WIB.
func validateClientAttendanceTime(
	clientDate string,
	clientClock string,
	serverTime time.Time,
) (time.Time, error) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.Time{}, ErrInvalidAttendanceTime
	}

	clientDate = strings.TrimSpace(clientDate)
	clientClock = strings.TrimSpace(clientClock)

	// Parse tanggal secara ketat.
	parsedDate, err := time.Parse("2006-01-02", clientDate)
	if err != nil {
		return time.Time{}, ErrInvalidAttendanceTime
	}

	// Parse jam secara ketat.
	parsedClock, err := time.Parse("15:04:05", clientClock)
	if err != nil {
		return time.Time{}, ErrInvalidAttendanceTime
	}

	clientTime := time.Date(
		parsedDate.Year(),
		parsedDate.Month(),
		parsedDate.Day(),
		parsedClock.Hour(),
		parsedClock.Minute(),
		parsedClock.Second(),
		0,
		loc,
	)

	diff := math.Abs(clientTime.Sub(serverTime).Seconds())
	if diff > attendanceTimeTolerance.Seconds() {
		return time.Time{}, ErrAttendanceTimeMismatch
	}

	return clientTime, nil
}
