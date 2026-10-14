package handler

import (
	"errors"
	"net/http"
	"strconv"

	"be-presensi/internal/modules/attendance/dto"
	"be-presensi/internal/modules/attendance/service"
	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/response"

	"github.com/gin-gonic/gin"
)

type AttendanceHandler struct {
	attendanceService *service.AttendanceService
}

func NewAttendanceHandler(
	attendanceService *service.AttendanceService,
) *AttendanceHandler {
	return &AttendanceHandler{
		attendanceService: attendanceService,
	}
}

// CheckIn godoc
// POST /api/attendances/check-in
func (h *AttendanceHandler) CheckIn(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(
			c,
			http.StatusUnauthorized,
			"invalid authenticated user",
		)
		return
	}

	var req dto.CheckInAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid check-in request",
		)
		return
	}

	attendance, err := h.attendanceService.CheckIn(
		c.Request.Context(),
		userID,
		req,
	)
	if err != nil {
		h.handleAttendanceError(c, err)
		return
	}

	response.Success(
		c,
		http.StatusCreated,
		"check-in successful",
		dto.ToAttendanceResponse(attendance),
	)
}

// CheckOut godoc
// POST /api/attendances/:id/check-out
func (h *AttendanceHandler) CheckOut(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(
			c,
			http.StatusUnauthorized,
			"invalid authenticated user",
		)
		return
	}

	attendanceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || attendanceID <= 0 {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid attendance id",
		)
		return
	}

	var req dto.CheckOutAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid check-out request",
		)
		return
	}

	attendance, err := h.attendanceService.CheckOut(
		c.Request.Context(),
		userID,
		attendanceID,
		req,
	)
	if err != nil {
		h.handleAttendanceError(c, err)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"check-out successful",
		dto.ToAttendanceResponse(attendance),
	)
}

func (h *AttendanceHandler) handleAttendanceError(
	c *gin.Context,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrInvalidAttendanceTime),
		errors.Is(err, service.ErrAttendanceTimeMismatch):
		response.Error(
			c,
			http.StatusBadRequest,
			"unable to process attendance request",
		)

	case errors.Is(err, service.ErrAttendanceUserNotFound),
		errors.Is(err, service.ErrAttendanceRoleNotFound),
		errors.Is(err, service.ErrAttendanceNotFound):
		response.Error(
			c,
			http.StatusNotFound,
			err.Error(),
		)

	case errors.Is(err, service.ErrAttendanceAlreadyExists),
		errors.Is(err, service.ErrAttendanceAlreadyCheckedOut),
		errors.Is(err, service.ErrAttendanceScheduleNotEnded),
		errors.Is(err, service.ErrDailyActivityRequired),
		errors.Is(err, service.ErrDailyActivityLocked),
		errors.Is(err, service.ErrAttendanceScheduleRequired):
		response.Error(
			c,
			http.StatusConflict,
			err.Error(),
		)

	case errors.Is(err, service.ErrInvalidCoordinates),
		errors.Is(err, service.ErrOutOfAreaNoteRequired):
		response.Error(
			c,
			http.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(err, service.ErrSchoolLocationUnavailable):
		response.Error(
			c,
			http.StatusServiceUnavailable,
			err.Error(),
		)

	default:
		// Detail error dicatat di server, bukan dibocorkan ke client.
		c.Error(err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to process attendance",
		)
	}
}
