package handler

import (
	"errors"
	"net/http"
	"strconv"

	"be-presensi/internal/modules/role/dto"
	"be-presensi/internal/modules/role/service"
	"be-presensi/internal/response"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	service service.RoleService
}

func NewRoleHandler(service service.RoleService) *RoleHandler {
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

	role, err := h.service.FindByID(c.Request.Context(), id)

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

	role, err := h.service.Create(c.Request.Context(), req)

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

	role, err := h.service.Update(
		c.Request.Context(),
		id,
		req,
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

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
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
		nil,
	)
}
