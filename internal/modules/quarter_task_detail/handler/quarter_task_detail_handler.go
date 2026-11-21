package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"be-presensi/internal/modules/auth/middleware"
	detailService "be-presensi/internal/modules/quarter_task_detail/service"
	"be-presensi/internal/response"

	quarterTaskDTO "be-presensi/internal/modules/quarter_task/dto"
	detailDTO "be-presensi/internal/modules/quarter_task_detail/dto"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type QuarterTaskDetailHandler struct {
	service *detailService.QuarterTaskDetailService
}

func NewQuarterTaskDetailHandler(
	quarterTaskDetailService *detailService.QuarterTaskDetailService,
) *QuarterTaskDetailHandler {
	return &QuarterTaskDetailHandler{
		service: quarterTaskDetailService,
	}
}

func (h *QuarterTaskDetailHandler) FindAllTasks(c *gin.Context) {
	tasks, err := h.service.FindAllTasks(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	data := make([]interface{}, 0, len(tasks))
	for _, task := range tasks {
		data = append(data, quarterTaskDTO.ToQuarterTaskResponse(task))
	}

	response.Success(
		c,
		http.StatusOK,
		"quarter tasks retrieved successfully",
		data,
	)
}

func (h *QuarterTaskDetailHandler) Upsert(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return
	}

	quarterTaskID, ok := parseQuarterTaskID(c)
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid quarter task id")
		return
	}

	var req detailDTO.UpsertQuarterTaskDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid quarter task detail request")
		return
	}

	detail, err := h.service.Upsert(
		c.Request.Context(),
		userID,
		quarterTaskID,
		&req,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"quarter task detail saved successfully",
		detailDTO.ToQuarterTaskDetailResponse(detail),
	)
}

func (h *QuarterTaskDetailHandler) GetByTask(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return
	}

	quarterTaskID, ok := parseQuarterTaskID(c)
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid quarter task id")
		return
	}

	detail, err := h.service.GetByTask(
		c.Request.Context(),
		userID,
		quarterTaskID,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"quarter task detail retrieved successfully",
		detailDTO.ToQuarterTaskDetailResponse(detail),
	)
}

func parseQuarterTaskID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	return id, err == nil && id > 0
}

func (h *QuarterTaskDetailHandler) handleError(
	c *gin.Context,
	err error,
) {
	switch {
	case errors.Is(err, detailService.ErrQuarterTaskNotFound),
		errors.Is(err, detailService.ErrQuarterTaskDetailNotFound),
		errors.Is(err, gorm.ErrRecordNotFound):
		response.Error(c, http.StatusNotFound, err.Error())

	case errors.Is(err, detailService.ErrQuarterTaskDetailLocked):
		response.Error(c, http.StatusConflict, err.Error())

	case errors.Is(err, detailService.ErrAchievementRequired),
		errors.Is(err, detailService.ErrInvalidQuarterTaskAction):
		response.Error(c, http.StatusBadRequest, err.Error())

	default:
		c.Error(err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to process quarter task detail",
		)
	}
}

func (h *QuarterTaskDetailHandler) FindAllSubmissionsPaginated(c *gin.Context) {
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

	var year *int

	if value := c.Query("year"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 2000 {
			response.Error(
				c,
				http.StatusBadRequest,
				"year must be a valid year",
			)
			return
		}
		year = &parsed
	}

	var quarterTaskID *int64

	if value := c.Query("quarter_task_id"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed < 1 {
			response.Error(
				c,
				http.StatusBadRequest,
				"quarter_task_id must be a positive integer",
			)
			return
		}
		quarterTaskID = &parsed
	}

	details, total, err := h.service.FindAllSubmissionsPaginated(
		c.Request.Context(),
		search,
		year,
		quarterTaskID,
		page,
		perPage,
	)

	if err != nil {
		c.Error(err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to get quarter task submissions",
		)
		return
	}

	data := make([]interface{}, 0, len(details))

	for _, detail := range details {
		data = append(
			data,
			detailDTO.ToQuarterTaskDetailResponse(detail),
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "quarter task submissions retrieved successfully",
		"data":    data,
		"meta": gin.H{
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
	})
}
