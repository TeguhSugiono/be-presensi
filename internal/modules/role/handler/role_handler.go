package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/role/dto"
	"be-presensi/internal/modules/role/service"
	"be-presensi/internal/response"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	service service.RoleService
}

func NewRoleHandler(
	service service.RoleService,
) *RoleHandler {
	return &RoleHandler{
		service: service,
	}
}

func (h *RoleHandler) FindAll(c *gin.Context) {
	roles, err := h.service.FindAll(c.Request.Context())

	if err != nil {
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to get roles",
		)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"roles retrieved successfully",
		roles,
	)
}

func (h *RoleHandler) FindAllPaginated(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))

	page := 1
	perPage := 10

	if value := c.Query("page"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			response.Error(c, http.StatusBadRequest,
				"page must be a positive integer")
			return
		}
		page = parsed
	}

	if value := c.Query("per_page"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			response.Error(c, http.StatusBadRequest,
				"per_page must be a positive integer")
			return
		}
		perPage = parsed
	}

	if perPage > 100 {
		perPage = 100
	}

	roles, total, err := h.service.FindAllPaginated(
		c.Request.Context(),
		search,
		page,
		perPage,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError,
			"failed to get roles")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "roles retrieved successfully",
		"data":    roles,
		"meta": gin.H{
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
	})
}

func (h *RoleHandler) FindByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid role id",
		)
		return
	}

	role, err := h.service.FindByID(
		c.Request.Context(),
		id,
	)

	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			response.Error(
				c,
				http.StatusNotFound,
				"role not found",
			)
			return
		}

		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to get role",
		)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"role retrieved successfully",
		role,
	)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleRequest

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

	role, err := h.service.Create(
		c.Request.Context(),
		&req,
		&userID,
	)

	if err != nil {
		if errors.Is(err, service.ErrRoleExists) {
			response.Error(
				c,
				http.StatusConflict,
				"role already exists",
			)
			return
		}

		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to create role",
		)
		return
	}

	response.Success(
		c,
		http.StatusCreated,
		"role created successfully",
		role,
	)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid role id",
		)
		return
	}

	var req dto.UpdateRoleRequest

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

	role, err := h.service.Update(
		c.Request.Context(),
		id,
		&req,
		&userID,
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoleNotFound):
			response.Error(
				c,
				http.StatusNotFound,
				"role not found",
			)

		case errors.Is(err, service.ErrRoleExists):
			response.Error(
				c,
				http.StatusConflict,
				"role already exists",
			)

		default:
			response.Error(
				c,
				http.StatusInternalServerError,
				"failed to update role",
			)
		}

		return
	}

	response.Success(
		c,
		http.StatusOK,
		"role updated successfully",
		role,
	)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid role id",
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
		if errors.Is(err, service.ErrRoleNotFound) {
			response.Error(
				c,
				http.StatusNotFound,
				"role not found",
			)
			return
		}

		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to delete role",
		)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"role deleted successfully",
		[]interface{}{},
	)
}
