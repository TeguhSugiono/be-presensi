package handler

import (
	"errors"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"be-presensi/internal/modules/attendance/dto"
	"be-presensi/internal/modules/attendance/service"
	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/response"

	attendanceStorage "be-presensi/internal/modules/attendance/storage"

	"github.com/gin-gonic/gin"
)

type AttendanceHandler struct {
	attendanceService *service.AttendanceService
	appBaseURL        string
}

func NewAttendanceHandler(
	attendanceService *service.AttendanceService,
	appBaseURL string,
) *AttendanceHandler {
	return &AttendanceHandler{
		attendanceService: attendanceService,
		appBaseURL:        strings.TrimRight(appBaseURL, "/"),
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
		//dto.ToAttendanceResponse(attendance),
		dto.ToAttendanceResponse(attendance, h.appBaseURL),
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

	// attendanceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	// if err != nil || attendanceID <= 0 {
	// 	response.Error(c, http.StatusBadRequest, "invalid attendance id")
	// 	return
	// }

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
		&req,
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
		//dto.ToAttendanceResponse(attendance),
		dto.ToAttendanceResponse(attendance, h.appBaseURL),
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
	// Validate attendance ID.
	attendanceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || attendanceID <= 0 {
		response.Error(c, http.StatusBadRequest, "invalid attendance id")
		return
	}

	// Retrieve photo path without authentication.
	photoPath, err := h.attendanceService.GetAttendancePhoto(
		c.Request.Context(),
		attendanceID,
		photoType,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAttendanceNotFound),
			errors.Is(err, service.ErrAttendancePhotoNotFound):
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

	// Only permit files inside the expected attendance photo directory.
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
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to retrieve attendance photo",
		)
		return
	}

	expectedAbsoluteDir, err := filepath.Abs(expectedDir)
	if err != nil {
		c.Error(err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to retrieve attendance photo",
		)
		return
	}

	relativePath, err := filepath.Rel(expectedAbsoluteDir, absolutePath)
	if err != nil ||
		relativePath == ".." ||
		strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		response.Error(c, http.StatusNotFound, "attendance photo not found")
		return
	}

	fileInfo, err := os.Stat(absolutePath)
	if err != nil || !fileInfo.Mode().IsRegular() {
		response.Error(c, http.StatusNotFound, "attendance photo not found")
		return
	}

	// Display image directly in browser.
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Type", "image/jpeg")
	c.File(absolutePath)
}

// FindAllPaginated godoc
// GET /api/attendances/paginated?page=1&per_page=10&search=

func (h *AttendanceHandler) FindAllPaginated(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return
	}

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

	// Optional date filter.
	// var date *time.Time

	// dateParam := strings.TrimSpace(c.Query("date"))

	// if dateParam != "" {
	// 	loc, err := time.LoadLocation("Asia/Jakarta")
	// 	if err != nil {
	// 		loc = time.FixedZone("WIB", 7*60*60)
	// 	}

	// 	parsedDate, err := time.ParseInLocation(
	// 		"2006-01-02",
	// 		dateParam,
	// 		loc,
	// 	)
	// 	if err != nil {
	// 		response.Error(
	// 			c,
	// 			http.StatusBadRequest,
	// 			"invalid date format, use YYYY-MM-DD",
	// 		)
	// 		return
	// 	}

	// 	date = &parsedDate
	// }

	// Optional date range filter.
	dateStartParam := strings.TrimSpace(c.Query("date_start"))
	dateEndParam := strings.TrimSpace(c.Query("date_end"))

	// Kedua parameter harus diisi bersama.
	if (dateStartParam == "") != (dateEndParam == "") {
		response.Error(
			c,
			http.StatusBadRequest,
			"date_start and date_end must be provided together",
		)
		return
	}

	var dateStart *time.Time
	var dateEnd *time.Time

	if dateStartParam != "" {
		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			loc = time.FixedZone("WIB", 7*60*60)
		}

		parsedStart, err := time.ParseInLocation(
			"2006-01-02",
			dateStartParam,
			loc,
		)
		if err != nil {
			response.Error(
				c,
				http.StatusBadRequest,
				"invalid date_start format, use YYYY-MM-DD",
			)
			return
		}

		parsedEnd, err := time.ParseInLocation(
			"2006-01-02",
			dateEndParam,
			loc,
		)
		if err != nil {
			response.Error(
				c,
				http.StatusBadRequest,
				"invalid date_end format, use YYYY-MM-DD",
			)
			return
		}

		if parsedStart.After(parsedEnd) {
			response.Error(
				c,
				http.StatusBadRequest,
				"date_start must be less than or equal to date_end",
			)
			return
		}

		dateStart = &parsedStart
		dateEnd = &parsedEnd
	}

	// attendances, total, err := h.attendanceService.FindAllPaginated(
	// 	c.Request.Context(),
	// 	userID,
	// 	search,
	// 	date,
	// 	page,
	// 	perPage,
	// )

	attendances, total, err := h.attendanceService.FindAllPaginated(
		c.Request.Context(),
		userID,
		search,
		dateStart,
		dateEnd,
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

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}

	emptyString := ""

	data := make([]dto.AttendanceListResponse, 0, len(attendances))

	for _, attendance := range attendances {
		item := dto.AttendanceListResponse{
			ID: attendance.ID,
			// UserName:           &attendance.UserName,
			UserName:           attendance.UserName,
			ScheduledStartTime: attendance.ScheduledStartTime,
			ScheduledEndTime:   attendance.ScheduledEndTime,

			CheckInPhoto:  &emptyString,
			KeteranganIn:  &emptyString,
			CheckOutPhoto: &emptyString,
			KeteranganOut: &emptyString,
			OutOfAreaNote: &emptyString,
		}

		// User ID.
		item.UserID = attendance.UserID

		// Attendance date.
		if attendance.AttendanceDate != nil {
			value := attendance.AttendanceDate.In(loc).Format("2006-01-02")
			item.AttendanceDate = &value
		}

		// Check-in information.
		if attendance.CheckInTime != nil {
			value := attendance.CheckInTime.In(loc).Format("2006-01-02 15:04:05")
			item.CheckInTime = &value
		}

		// if attendance.CheckInPhoto != nil {
		// 	item.CheckInPhoto = attendance.CheckInPhoto
		// }

		if attendance.CheckInPhoto != nil &&
			attendance.ID != nil &&
			*attendance.ID > 0 {
			photoURL := h.appBaseURL +
				"/api/attendances/" +
				strconv.FormatInt(*attendance.ID, 10) +
				"/check-in-photo"

			item.CheckInPhoto = &photoURL
		}

		item.CheckInLat = attendance.CheckInLat
		item.CheckInLng = attendance.CheckInLng

		if attendance.KeteranganIn != nil {
			item.KeteranganIn = attendance.KeteranganIn
		}

		item.IsLate = attendance.IsLate

		// Check-out information.
		if attendance.CheckOutTime != nil {
			value := attendance.CheckOutTime.In(loc).Format("2006-01-02 15:04:05")
			item.CheckOutTime = &value
		}

		// if attendance.CheckOutPhoto != nil {
		// 	item.CheckOutPhoto = attendance.CheckOutPhoto
		// }

		if attendance.CheckOutPhoto != nil &&
			attendance.ID != nil &&
			*attendance.ID > 0 {
			photoURL := h.appBaseURL +
				"/api/attendances/" +
				strconv.FormatInt(*attendance.ID, 10) +
				"/check-out-photo"

			item.CheckOutPhoto = &photoURL
		}

		item.CheckOutLat = attendance.CheckOutLat
		item.CheckOutLng = attendance.CheckOutLng

		if attendance.KeteranganOut != nil {
			item.KeteranganOut = attendance.KeteranganOut
		}

		item.IsOutOfArea = attendance.IsOutOfArea

		if attendance.OutOfAreaNote != nil {
			item.OutOfAreaNote = attendance.OutOfAreaNote
		}

		// Audit information.
		if attendance.CreatedAt != nil {
			value := attendance.CreatedAt.In(loc).Format("2006-01-02 15:04:05")
			item.CreatedAt = &value
		}

		if attendance.UpdatedAt != nil {
			value := attendance.UpdatedAt.In(loc).Format("2006-01-02 15:04:05")
			item.UpdatedAt = &value
		}

		data = append(data, item)
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

	case errors.Is(err, service.ErrAttendanceTimeMismatch):
		response.Error(
			c,
			http.StatusBadRequest,
			"Waktu yang anda kirim, tidak sesuai dengan waktu saat ini, Silahkan untuk lakukan absen kembali",
		)

	case errors.Is(err, service.ErrInvalidAttendanceTime):
		response.Error(
			c,
			http.StatusBadRequest,
			"Format tanggal atau waktu absensi tidak valid",
		)

	case errors.Is(err, service.ErrAttendanceUserNotFound),
		errors.Is(err, service.ErrAttendanceRoleNotFound),
		errors.Is(err, service.ErrAttendanceNotFound):
		response.Error(
			c,
			http.StatusNotFound,
			err.Error(),
		)

	case errors.Is(err, service.ErrDailyActivityRequired),
		errors.Is(err, service.ErrDailyActivityNotSubmitted):
		response.Error(
			c,
			http.StatusBadRequest,
			"Silakan submit Daily Activity terlebih dahulu sebelum melakukan checkout.",
		)

	case errors.Is(err, service.ErrAttendanceAlreadyExists),
		errors.Is(err, service.ErrAttendanceAlreadyCheckedOut),
		errors.Is(err, service.ErrAttendanceScheduleNotEnded),
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

func (h *AttendanceHandler) CheckInDetail(c *gin.Context) {
	h.getAttendanceDetail(c, "check-in")
}

func (h *AttendanceHandler) CheckOutDetail(c *gin.Context) {
	h.getAttendanceDetail(c, "check-out")
}

func (h *AttendanceHandler) getAttendanceDetail(
	c *gin.Context,
	detailType string,
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

	attendance, err := h.attendanceService.GetAttendanceDetail(
		c.Request.Context(),
		userID,
		attendanceID,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAttendanceNotFound),
			errors.Is(err, service.ErrAttendanceUserNotFound),
			errors.Is(err, service.ErrAttendanceRoleNotFound):
			response.Error(c, http.StatusNotFound, "attendance not found")
		default:
			c.Error(err)
			response.Error(c, http.StatusInternalServerError, "failed to retrieve attendance detail")
		}
		return
	}

	switch detailType {
	case "check-in":
		response.Success(
			c,
			http.StatusOK,
			"check-in detail retrieved successfully",
			dto.AttendanceCheckInDetailResponse{
				ID:             attendance.ID,
				AttendanceDate: attendance.AttendanceDate.Format("2006-01-02"),
				CheckInTime:    attendance.CheckInTime.Format("2006-01-02 15:04:05"),
				PhotoURL:       h.appBaseURL + "/api/attendances/" + strconv.FormatInt(attendance.ID, 10) + "/check-in-photo",
				Latitude:       attendance.CheckInLat,
				Longitude:      attendance.CheckInLng,
				Keterangan:     attendance.KeteranganIn,
			},
		)

	case "check-out":
		if attendance.CheckOutTime == nil ||
			attendance.CheckOutPhoto == nil ||
			attendance.CheckOutLat == nil ||
			attendance.CheckOutLng == nil {
			response.Error(c, http.StatusNotFound, "check-out detail not found")
			return
		}

		response.Success(
			c,
			http.StatusOK,
			"check-out detail retrieved successfully",
			dto.AttendanceCheckOutDetailResponse{
				ID:             attendance.ID,
				AttendanceDate: attendance.AttendanceDate.Format("2006-01-02"),
				CheckOutTime:   attendance.CheckOutTime.Format("2006-01-02 15:04:05"),
				PhotoURL:       h.appBaseURL + "/api/attendances/" + strconv.FormatInt(attendance.ID, 10) + "/check-out-photo",
				Latitude:       *attendance.CheckOutLat,
				Longitude:      *attendance.CheckOutLng,
				Keterangan:     attendance.KeteranganOut,
			},
		)
	}
}

// GetTodayAttendance godoc
// GET /api/attendances/today
func (h *AttendanceHandler) GetTodayAttendance(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return
	}

	attendance, err := h.attendanceService.GetTodayAttendance(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAttendanceNotFound):
			response.Error(c, http.StatusNotFound, "attendance not found")

		case errors.Is(err, service.ErrAttendanceUserNotFound),
			errors.Is(err, service.ErrAttendanceRoleNotFound):
			response.Error(c, http.StatusNotFound, err.Error())

		default:
			c.Error(err)
			response.Error(
				c,
				http.StatusInternalServerError,
				"failed to retrieve today's attendance",
			)
		}
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"attendance detail retrieved successfully",
		dto.ToAttendanceTodayResponse(attendance, h.appBaseURL),
	)
}

func (h *AttendanceHandler) GetSummary(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}

	dateParam := strings.TrimSpace(c.Query("date"))

	var date time.Time

	if dateParam == "" {
		date = time.Now().In(loc)
	} else {
		parsedDate, err := time.ParseInLocation(
			"2006-01-02",
			dateParam,
			loc,
		)
		if err != nil {
			response.Error(
				c,
				http.StatusBadRequest,
				"invalid date format, use YYYY-MM-DD",
			)
			return
		}

		date = parsedDate
	}

	summary, err := h.attendanceService.GetSummary(
		c.Request.Context(),
		date,
	)
	if err != nil {
		c.Error(err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to retrieve attendance summary",
		)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"attendance summary retrieved successfully",
		summary,
	)
}
