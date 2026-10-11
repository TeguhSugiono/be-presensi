package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/user/dto"
	"be-presensi/internal/modules/user/service"
	"be-presensi/internal/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) FindAll(c *gin.Context) {
	users, err := h.service.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to get users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "users retrieved successfully",
		"data":    users,
	})
}

func (h *UserHandler) FindAllPaginated(c *gin.Context) {
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

	users, total, err := h.service.FindAllPaginated(
		search,
		page,
		perPage,
	)
	if err != nil {
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to get users",
		)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "users retrieved successfully",
		"data":    users,
		"meta": gin.H{
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
	})
}

func (h *UserHandler) FindByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid user id",
		})
		return
	}

	user, err := h.service.FindByID(id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "user not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to get user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user retrieved successfully",
		"data":    user,
	})
}

func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Sementara belum ada authentication.
	// Setelah modul Authorization selesai,
	// userID akan diambil dari JWT/session.
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	user, err := h.service.Create(req, &userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoleNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "role not found or inactive",
			})
		case errors.Is(err, service.ErrScheduleNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "work schedule not found or inactive",
			})
		case errors.Is(err, service.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "email already exists",
			})
		case errors.Is(err, service.ErrNIPAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "nip already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to create user",
			})
		}

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "user created successfully",
		"data":    user,
	})
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid user id",
		})
		return
	}

	var req dto.UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	user, err := h.service.Update(id, req, &userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "user not found",
			})
		case errors.Is(err, service.ErrRoleNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "role not found or inactive",
			})
		case errors.Is(err, service.ErrScheduleNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "work schedule not found or inactive",
			})
		case errors.Is(err, service.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "email already exists",
			})
		case errors.Is(err, service.ErrNIPAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "nip already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to update user",
			})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user updated successfully",
		"data":    user,
	})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid user id",
		})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	err = h.service.Delete(id, &userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "user not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to delete user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user deleted successfully",
		"data":    []interface{}{},
	})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid user id",
		)
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid password request",
		)
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		response.Error(
			c,
			http.StatusBadRequest,
			"password confirmation does not match",
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

	err = h.service.ChangePassword(
		id,
		req.NewPassword,
		&userID,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			response.Error(
				c,
				http.StatusNotFound,
				"user not found",
			)

		case errors.Is(err, service.ErrAdminRequired):
			response.Error(
				c,
				http.StatusForbidden,
				"admin access required",
			)

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to change password",
			})
		}
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"user password changed successfully",
		[]interface{}{},
	)

}
