package handler

import (
	"errors"
	"net/http"
	"strconv"

	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/work_schedule/dto"
	"be-presensi/internal/modules/work_schedule/service"
	"be-presensi/internal/response"

	"github.com/gin-gonic/gin"
)

type WorkScheduleHandler struct {
	service service.WorkScheduleService
}

func NewWorkScheduleHandler(
	service service.WorkScheduleService,
) *WorkScheduleHandler {
	return &WorkScheduleHandler{
		service: service,
	}
}

func (h *WorkScheduleHandler) FindAll(c *gin.Context) {
	schedules, err := h.service.FindAll(c.Request.Context())

	if err != nil {
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to get work schedules",
		)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"work schedules retrieved successfully",
		schedules,
	)
}

func (h *WorkScheduleHandler) FindByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid work schedule id",
		)
		return
	}

	schedule, err := h.service.FindByID(
		c.Request.Context(),
		id,
	)

	if err != nil {
		if errors.Is(err, service.ErrWorkScheduleNotFound) {
			response.Error(
				c,
				http.StatusNotFound,
				"work schedule not found",
			)
			return
		}

		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to get work schedule",
		)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"work schedule retrieved successfully",
		schedule,
	)
}

func (h *WorkScheduleHandler) Create(c *gin.Context) {
	var req dto.CreateWorkScheduleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid request",
		)
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(
			c,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	schedule, err := h.service.Create(
		c.Request.Context(),
		&req,
		&userID,
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrWorkScheduleExists):
			response.Error(
				c,
				http.StatusConflict,
				"work schedule already exists",
			)

		case errors.Is(err, service.ErrInvalidStartTime):
			response.Error(
				c,
				http.StatusBadRequest,
				"invalid start time, use HH:mm format",
			)

		case errors.Is(err, service.ErrInvalidEndTime):
			response.Error(
				c,
				http.StatusBadRequest,
				"invalid end time, use HH:mm format",
			)

		case errors.Is(err, service.ErrInvalidTimeRange):
			response.Error(
				c,
				http.StatusBadRequest,
				"start time must be before end time",
			)

		default:
			response.Error(
				c,
				http.StatusInternalServerError,
				"failed to create work schedule",
			)
		}

		return
	}

	response.Success(
		c,
		http.StatusCreated,
		"work schedule created successfully",
		schedule,
	)
}

func (h *WorkScheduleHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid work schedule id",
		)
		return
	}

	var req dto.UpdateWorkScheduleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid request",
		)
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(
			c,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	schedule, err := h.service.Update(
		c.Request.Context(),
		id,
		&req,
		&userID,
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrWorkScheduleNotFound):
			response.Error(
				c,
				http.StatusNotFound,
				"work schedule not found",
			)

		case errors.Is(err, service.ErrWorkScheduleExists):
			response.Error(
				c,
				http.StatusConflict,
				"work schedule already exists",
			)

		case errors.Is(err, service.ErrInvalidStartTime):
			response.Error(
				c,
				http.StatusBadRequest,
				"invalid start time, use HH:mm format",
			)

		case errors.Is(err, service.ErrInvalidEndTime):
			response.Error(
				c,
				http.StatusBadRequest,
				"invalid end time, use HH:mm format",
			)

		case errors.Is(err, service.ErrInvalidTimeRange):
			response.Error(
				c,
				http.StatusBadRequest,
				"start time must be before end time",
			)

		default:
			response.Error(
				c,
				http.StatusInternalServerError,
				"failed to update work schedule",
			)
		}

		return
	}

	response.Success(
		c,
		http.StatusOK,
		"work schedule updated successfully",
		schedule,
	)
}

func (h *WorkScheduleHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid work schedule id",
		)
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(
			c,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	if err := h.service.Delete(
		c.Request.Context(),
		id,
		&userID,
	); err != nil {

		if errors.Is(err, service.ErrWorkScheduleNotFound) {
			response.Error(
				c,
				http.StatusNotFound,
				"work schedule not found",
			)
			return
		}

		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to delete work schedule",
		)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"work schedule deleted successfully",
		[]interface{}{},
	)
}
