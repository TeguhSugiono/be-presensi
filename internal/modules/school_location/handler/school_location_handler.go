package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/school_location/dto"
	"be-presensi/internal/modules/school_location/service"
	"be-presensi/internal/response"

	"github.com/gin-gonic/gin"
)

type SchoolLocationHandler struct {
	service service.SchoolLocationService
}

func NewSchoolLocationHandler(
	service service.SchoolLocationService,
) *SchoolLocationHandler {
	return &SchoolLocationHandler{
		service: service,
	}
}

func (h *SchoolLocationHandler) FindAll(c *gin.Context) {
	locations, err := h.service.FindAll(c.Request.Context())
	if err != nil {
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to get school locations",
		)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"school locations retrieved successfully",
		locations,
	)
}

func (h *SchoolLocationHandler) FindAllPaginated(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))

	page := 1
	if value := c.Query("page"); value != "" {
		parsedPage, err := strconv.Atoi(value)
		if err != nil || parsedPage < 1 {
			response.Error(
				c,
				http.StatusBadRequest,
				"page must be a positive integer",
			)
			return
		}
		page = parsedPage
	}

	perPage := 10
	if value := c.Query("per_page"); value != "" {
		parsedPerPage, err := strconv.Atoi(value)
		if err != nil || parsedPerPage < 1 {
			response.Error(
				c,
				http.StatusBadRequest,
				"per_page must be a positive integer",
			)
			return
		}
		perPage = parsedPerPage
	}

	if perPage > 100 {
		perPage = 100
	}

	locations, total, err := h.service.FindAllPaginated(
		c.Request.Context(),
		search,
		page,
		perPage,
	)
	if err != nil {
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to get school locations",
		)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "school locations retrieved successfully",
		"data":    locations,
		"meta": gin.H{
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
	})
}

func (h *SchoolLocationHandler) FindByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid school location id",
		)
		return
	}

	location, err := h.service.FindByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		if errors.Is(err, service.ErrSchoolLocationNotFound) {
			response.Error(
				c,
				http.StatusNotFound,
				"school location not found",
			)
			return
		}

		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to get school location",
		)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"school location retrieved successfully",
		location,
	)
}

func (h *SchoolLocationHandler) Create(c *gin.Context) {
	var req dto.CreateSchoolLocationRequest

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

	location, err := h.service.Create(
		c.Request.Context(),
		&req,
		&userID,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSchoolLocationExists):
			response.Error(
				c,
				http.StatusConflict,
				"school location already exists",
			)
		case errors.Is(err, service.ErrInvalidLocationName):
			response.Error(
				c,
				http.StatusBadRequest,
				"location name is required",
			)
		case errors.Is(err, service.ErrInvalidLatitude):
			response.Error(
				c,
				http.StatusBadRequest,
				"latitude must be between -90 and 90",
			)
		case errors.Is(err, service.ErrInvalidLongitude):
			response.Error(
				c,
				http.StatusBadRequest,
				"longitude must be between -180 and 180",
			)
		case errors.Is(err, service.ErrInvalidRadius):
			response.Error(
				c,
				http.StatusBadRequest,
				"radius must be greater than zero",
			)
		default:
			response.Error(
				c,
				http.StatusInternalServerError,
				"failed to create school location",
			)
		}
		return
	}

	response.Success(
		c,
		http.StatusCreated,
		"school location created successfully",
		location,
	)
}

func (h *SchoolLocationHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid school location id",
		)
		return
	}

	var req dto.UpdateSchoolLocationRequest

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

	location, err := h.service.Update(
		c.Request.Context(),
		id,
		&req,
		&userID,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSchoolLocationNotFound):
			response.Error(
				c,
				http.StatusNotFound,
				"school location not found",
			)
		case errors.Is(err, service.ErrSchoolLocationExists):
			response.Error(
				c,
				http.StatusConflict,
				"school location already exists",
			)
		case errors.Is(err, service.ErrInvalidLocationName):
			response.Error(
				c,
				http.StatusBadRequest,
				"location name is required",
			)
		case errors.Is(err, service.ErrInvalidLatitude):
			response.Error(
				c,
				http.StatusBadRequest,
				"latitude must be between -90 and 90",
			)
		case errors.Is(err, service.ErrInvalidLongitude):
			response.Error(
				c,
				http.StatusBadRequest,
				"longitude must be between -180 and 180",
			)
		case errors.Is(err, service.ErrInvalidRadius):
			response.Error(
				c,
				http.StatusBadRequest,
				"radius must be greater than zero",
			)
		default:
			response.Error(
				c,
				http.StatusInternalServerError,
				"failed to update school location",
			)
		}
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"school location updated successfully",
		location,
	)
}

func (h *SchoolLocationHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid school location id",
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
		if errors.Is(err, service.ErrSchoolLocationNotFound) {
			response.Error(
				c,
				http.StatusNotFound,
				"school location not found",
			)
			return
		}

		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to delete school location",
		)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"school location deleted successfully",
		[]interface{}{},
	)
}
