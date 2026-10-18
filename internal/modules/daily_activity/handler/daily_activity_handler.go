package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/daily_activity/dto"
	"be-presensi/internal/modules/daily_activity/service"
	"be-presensi/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DailyActivityHandler struct {
	service *service.DailyActivityService
}

func NewDailyActivityHandler(
	dailyActivityService *service.DailyActivityService,
) *DailyActivityHandler {
	return &DailyActivityHandler{
		service: dailyActivityService,
	}
}

func (h *DailyActivityHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return
	}

	attendanceID, ok := parseAttendanceID(c)
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid attendance id")
		return
	}

	var req dto.CreateDailyActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid daily activity request")
		return
	}

	activity, err := h.service.Create(
		c.Request.Context(),
		userID,
		attendanceID,
		req,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(
		c,
		http.StatusCreated,
		"daily activity created successfully",
		dto.ToDailyActivityResponse(activity),
	)
}

func (h *DailyActivityHandler) GetByAttendance(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return
	}

	attendanceID, ok := parseAttendanceID(c)
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid attendance id")
		return
	}

	activity, err := h.service.GetByAttendance(
		c.Request.Context(),
		userID,
		attendanceID,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"daily activity retrieved successfully",
		dto.ToDailyActivityResponse(activity),
	)
}

func (h *DailyActivityHandler) GetByDate(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(
			c,
			http.StatusUnauthorized,
			"invalid authenticated user",
		)
		return
	}

	dateStr := strings.TrimSpace(
		c.Query("activity_date"),
	)

	if dateStr == "" {
		response.Error(
			c,
			http.StatusBadRequest,
			"activity_date is required",
		)
		return
	}

	activityDate, err := time.Parse(
		"2006-01-02",
		dateStr,
	)
	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid activity_date format, use YYYY-MM-DD",
		)
		return
	}

	activity, err := h.service.GetByDate(
		c.Request.Context(),
		userID,
		activityDate,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"daily activity retrieved successfully",
		dto.ToDailyActivityResponse(activity),
	)
}

func (h *DailyActivityHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return
	}

	attendanceID, ok := parseAttendanceID(c)
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid attendance id")
		return
	}

	var req dto.UpdateDailyActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid daily activity request")
		return
	}

	activity, err := h.service.Update(
		c.Request.Context(),
		userID,
		attendanceID,
		req,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"daily activity updated successfully",
		dto.ToDailyActivityResponse(activity),
	)
}

func parseAttendanceID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("attendance_id"), 10, 64)
	return id, err == nil && id > 0
}

func (h *DailyActivityHandler) handleError(
	c *gin.Context,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrAttendanceNotFound),
		errors.Is(err, service.ErrActivityNotFound):
		response.Error(c, http.StatusNotFound, err.Error())

	case errors.Is(err, service.ErrActivityAlreadyExists),
		errors.Is(err, service.ErrAttendanceAlreadyChecked),
		errors.Is(err, service.ErrActivityLocked):
		response.Error(c, http.StatusConflict, err.Error())

	case errors.Is(err, service.ErrDescriptionRequired):
		response.Error(c, http.StatusBadRequest, err.Error())

	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Error(c, http.StatusNotFound, "daily activity not found")

	default:
		c.Error(err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to process daily activity",
		)
	}
}

func (h *DailyActivityHandler) Upsert(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(
			c,
			http.StatusUnauthorized,
			"invalid authenticated user",
		)
		return
	}

	attendanceID, ok := parseAttendanceID(c)
	if !ok {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid attendance id",
		)
		return
	}

	var req dto.UpsertDailyActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	activity, err := h.service.Upsert(
		c.Request.Context(),
		userID,
		attendanceID,
		req,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"daily activity saved successfully",
		dto.ToDailyActivityResponse(activity),
	)
}

func (h *DailyActivityHandler) FindAllPaginated(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))

	page := 1
	perPage := 10

	if value := c.Query("page"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			response.Error(
				c,
				http.StatusBadRequest,
				"page must be a positive integer",
			)
			return
		}
		page = parsed
	}

	if value := c.Query("per_page"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			response.Error(
				c,
				http.StatusBadRequest,
				"per_page must be a positive integer",
			)
			return
		}
		perPage = parsed
	}

	if perPage > 100 {
		perPage = 100
	}

	activities, total, err := h.service.FindAllPaginated(
		search,
		page,
		perPage,
	)
	if err != nil {
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to get daily activities",
		)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "daily activities retrieved successfully",
		"data":    activities,
		"meta": gin.H{
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
	})
}
