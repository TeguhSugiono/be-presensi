package handler

import (
	"errors"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"be-presensi/internal/modules/attendance/dto"
	"be-presensi/internal/modules/attendance/service"
	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/response"

	attendanceStorage "be-presensi/internal/modules/attendance/storage"

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
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return
	}

	var req dto.CheckInAttendanceRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid check-in request")
		return
	}

	fileHeader, err := c.FormFile("check_in_photo")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "check-in photo is required")
		return
	}

	photoPath, err := attendanceStorage.SaveAttendancePhoto(
		fileHeader,
		"check-in",
	)
	if err != nil {
		h.handlePhotoUploadError(c, err)
		return
	}

	req.CheckInPhoto = photoPath

	attendance, err := h.attendanceService.CheckIn(
		c.Request.Context(),
		userID,
		req,
	)
	if err != nil {
		_ = os.Remove(photoPath)
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
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return
	}

	attendanceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || attendanceID <= 0 {
		response.Error(c, http.StatusBadRequest, "invalid attendance id")
		return
	}

	var req dto.CheckOutAttendanceRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid check-out request")
		return
	}

	fileHeader, err := c.FormFile("check_out_photo")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "check-out photo is required")
		return
	}

	photoPath, err := attendanceStorage.SaveAttendancePhoto(
		fileHeader,
		"check-out",
	)
	if err != nil {
		h.handlePhotoUploadError(c, err)
		return
	}

	req.CheckOutPhoto = photoPath

	attendance, err := h.attendanceService.CheckOut(
		c.Request.Context(),
		userID,
		attendanceID,
		req,
	)
	if err != nil {
		_ = os.Remove(photoPath)
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

func (h *AttendanceHandler) CheckInPhoto(c *gin.Context) {
	h.serveAttendancePhoto(c, "check-in")
}

func (h *AttendanceHandler) CheckOutPhoto(c *gin.Context) {
	h.serveAttendancePhoto(c, "check-out")
}

func (h *AttendanceHandler) serveAttendancePhoto(
	c *gin.Context,
	photoType string,
) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return
	}

	attendanceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || attendanceID <= 0 {
		response.Error(c, http.StatusBadRequest, "invalid attendance id")
		return
	}

	photoPath, err := h.attendanceService.GetAttendancePhoto(
		c.Request.Context(),
		userID,
		attendanceID,
		photoType,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAttendanceNotFound),
			errors.Is(err, service.ErrAttendancePhotoNotFound),
			errors.Is(err, service.ErrAttendanceUserNotFound):
			response.Error(c, http.StatusNotFound, "attendance photo not found")
		default:
			c.Error(err)
			response.Error(
				c,
				http.StatusInternalServerError,
				"failed to retrieve attendance photo",
			)
		}
		return
	}

	// Only permit files inside the expected private photo directory.
	expectedDir := filepath.Join("assets", "attendance", photoType)
	cleanPath := filepath.Clean(photoPath)

	if filepath.IsAbs(cleanPath) ||
		cleanPath != expectedDir &&
			!strings.HasPrefix(cleanPath, expectedDir+string(filepath.Separator)) {
		response.Error(c, http.StatusNotFound, "attendance photo not found")
		return
	}

	absolutePath, err := filepath.Abs(cleanPath)
	if err != nil {
		c.Error(err)
		response.Error(c, http.StatusInternalServerError, "failed to retrieve attendance photo")
		return
	}

	expectedAbsoluteDir, err := filepath.Abs(expectedDir)
	if err != nil {
		c.Error(err)
		response.Error(c, http.StatusInternalServerError, "failed to retrieve attendance photo")
		return
	}

	relativePath, err := filepath.Rel(expectedAbsoluteDir, absolutePath)
	if err != nil || relativePath == ".." ||
		strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		response.Error(c, http.StatusNotFound, "attendance photo not found")
		return
	}

	fileInfo, err := os.Stat(absolutePath)
	if err != nil || !fileInfo.Mode().IsRegular() {
		response.Error(c, http.StatusNotFound, "attendance photo not found")
		return
	}

	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Type", "image/jpeg")
	c.File(absolutePath)
}

// FindAllPaginated godoc
// GET /api/attendances/paginated?page=1&per_page=10&search=
func (h *AttendanceHandler) FindAllPaginated(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		response.Error(c, http.StatusBadRequest, "invalid page")
		return
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if err != nil || perPage < 1 {
		response.Error(c, http.StatusBadRequest, "invalid per_page")
		return
	}

	if perPage > 100 {
		perPage = 100
	}

	search := strings.TrimSpace(c.Query("search"))

	attendances, total, err := h.attendanceService.FindAllPaginated(
		c.Request.Context(),
		search,
		page,
		perPage,
	)
	if err != nil {
		c.Error(err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to retrieve attendances",
		)
		return
	}

	data := make([]dto.AttendanceListResponse, 0, len(attendances))

	for _, attendance := range attendances {
		item := dto.ToAttendanceResponse(attendance)

		data = append(data, dto.AttendanceListResponse{
			ID:                 item.ID,
			UserID:             item.UserID,
			UserName:           attendance.UserName,
			AttendanceDate:     item.AttendanceDate,
			ScheduledStartTime: item.ScheduledStartTime,
			ScheduledEndTime:   item.ScheduledEndTime,
			CheckInTime:        item.CheckInTime,
			CheckInPhoto:       item.CheckInPhoto,
			CheckInLat:         item.CheckInLat,
			CheckInLng:         item.CheckInLng,
			IsLate:             item.IsLate,
			CheckOutTime:       item.CheckOutTime,
			CheckOutPhoto:      item.CheckOutPhoto,
			CheckOutLat:        item.CheckOutLat,
			CheckOutLng:        item.CheckOutLng,
			IsOutOfArea:        item.IsOutOfArea,
			OutOfAreaNote:      item.OutOfAreaNote,
			CreatedAt:          item.CreatedAt,
			UpdatedAt:          item.UpdatedAt,
		})
	}

	response.Success(
		c,
		http.StatusOK,
		"attendances retrieved successfully",
		gin.H{
			"items": data,
			"pagination": gin.H{
				"page":        page,
				"per_page":    perPage,
				"total":       total,
				"total_pages": int(math.Ceil(float64(total) / float64(perPage))),
			},
		},
	)
}

func (h *AttendanceHandler) handlePhotoUploadError(
	c *gin.Context,
	err error,
) {
	switch {
	case errors.Is(err, attendanceStorage.ErrPhotoRequired),
		errors.Is(err, attendanceStorage.ErrInvalidPhoto),
		errors.Is(err, attendanceStorage.ErrInvalidPhotoType),
		errors.Is(err, attendanceStorage.ErrPhotoCompression):
		response.Error(
			c,
			http.StatusBadRequest,
			err.Error(),
		)

	default:
		c.Error(err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to process attendance photo",
		)
	}
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
