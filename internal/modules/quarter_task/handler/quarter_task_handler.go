package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/quarter_task/dto"
	"be-presensi/internal/modules/quarter_task/service"
	quarterTaskDetailDTO "be-presensi/internal/modules/quarter_task_detail/dto"
	"be-presensi/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type QuarterTaskHandler struct {
	service *service.QuarterTaskService
}

func NewQuarterTaskHandler(
	quarterTaskService *service.QuarterTaskService,
) *QuarterTaskHandler {
	return &QuarterTaskHandler{
		service: quarterTaskService,
	}
}

func (h *QuarterTaskHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return
	}

	var req dto.CreateQuarterTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid quarter task request")
		return
	}

	task, err := h.service.Create(
		c.Request.Context(),
		userID,
		&req,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(
		c,
		http.StatusCreated,
		"quarter task created successfully",
		dto.ToQuarterTaskResponse(task),
	)
}

func (h *QuarterTaskHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return
	}

	taskID, ok := parseQuarterTaskID(c)
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid quarter task id")
		return
	}

	var req dto.UpdateQuarterTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid quarter task request")
		return
	}

	task, err := h.service.Update(
		c.Request.Context(),
		userID,
		taskID,
		&req,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"quarter task updated successfully",
		dto.ToQuarterTaskResponse(task),
	)
}

func (h *QuarterTaskHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return
	}

	taskID, ok := parseQuarterTaskID(c)
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid quarter task id")
		return
	}

	err := h.service.Delete(
		c.Request.Context(),
		userID,
		taskID,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"quarter task deleted successfully",
		[]interface{}{},
	)
}

func (h *QuarterTaskHandler) GetByID(c *gin.Context) {
	taskID, ok := parseQuarterTaskID(c)
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid quarter task id")
		return
	}

	task, err := h.service.FindByID(
		c.Request.Context(),
		taskID,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"quarter task retrieved successfully",
		dto.ToQuarterTaskResponse(task),
	)
}

func (h *QuarterTaskHandler) FindAllPaginated(c *gin.Context) {
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

	// tasks, total, err := h.service.FindAllPaginated(
	// 	c.Request.Context(),
	// 	search,
	// 	year,
	// 	page,
	// 	perPage,
	// )

	tasks, total, err := h.service.FindAllPaginated(
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
			"failed to get quarter tasks",
		)
		return
	}

	data := make([]interface{}, 0, len(tasks))
	for _, task := range tasks {
		data = append(data, dto.ToQuarterTaskResponse(task))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "quarter tasks retrieved successfully",
		"data":    data,
		"meta": gin.H{
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
	})
}

func (h *QuarterTaskHandler) GetSubmissions(c *gin.Context) {
	taskID, ok := parseQuarterTaskID(c)
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid quarter task id")
		return
	}

	submissions, err := h.service.FindSubmissions(
		c.Request.Context(),
		taskID,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	data := make([]interface{}, 0, len(submissions))
	for _, submission := range submissions {
		data = append(
			data,
			quarterTaskDetailDTO.ToQuarterTaskDetailResponse(submission),
		)
	}

	response.Success(
		c,
		http.StatusOK,
		"quarter task submissions retrieved successfully",
		data,
	)
}

func parseQuarterTaskID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	return id, err == nil && id > 0
}

func (h *QuarterTaskHandler) handleError(
	c *gin.Context,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrQuarterTaskNotFound),
		errors.Is(err, gorm.ErrRecordNotFound):
		response.Error(c, http.StatusNotFound, err.Error())

	case errors.Is(err, service.ErrQuarterTaskHasSubmission):
		response.Error(c, http.StatusConflict, err.Error())

	case errors.Is(err, service.ErrQuarterTaskTitleRequired),
		errors.Is(err, service.ErrQuarterTaskInvalidDate),
		errors.Is(err, service.ErrQuarterTaskInvalidPeriod):
		response.Error(c, http.StatusBadRequest, err.Error())

	default:
		c.Error(err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to process quarter task",
		)
	}
}
