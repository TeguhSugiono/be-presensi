package handler

import (
	"errors"
	"net/http"
	"strconv"

	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/user/dto"
	"be-presensi/internal/modules/user/service"

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
