package handler

import (
	"errors"
	"net/http"
	"strconv"

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
