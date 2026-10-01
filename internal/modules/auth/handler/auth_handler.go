package handler

import (
	"errors"
	"net/http"

	"be-presensi/internal/modules/auth/dto"
	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/auth/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.Login(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid email or password",
			})

		case errors.Is(err, service.ErrUserInactive):
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "user is inactive",
			})

		case errors.Is(err, service.ErrUserDeleted):
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid email or password",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to login",
			})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "login successful",
		"data":    result,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	user, err := h.service.GetCurrentUser(userID)
	if err != nil {
		if errors.Is(err, service.ErrUserDeleted) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "user not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to get current user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "current user retrieved successfully",
		"data":    user,
	})
}

func (h *AuthHandler) SetupAdmin(c *gin.Context) {
	var req dto.SetupAdminRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.SetupAdmin(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "admin already exists",
			})

		case errors.Is(err, service.ErrAdminRoleNotFound):
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "admin role not found",
			})

		case errors.Is(err, service.ErrAdminEmailExists):
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "email already exists",
			})

		case errors.Is(err, service.ErrAdminNIPExists):
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "nip already exists",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to setup admin",
			})
		}

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "admin setup successfully",
		"data":    result,
	})
}
