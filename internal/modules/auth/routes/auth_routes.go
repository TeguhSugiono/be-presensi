package routes

import (
	"be-presensi/internal/modules/auth/handler"
	"be-presensi/internal/modules/auth/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.RouterGroup, handler *handler.AuthHandler) {
	auth := router.Group("/auth")

	auth.POST("/login", handler.Login)

	auth.GET("/me", middleware.AuthMiddleware(), handler.Me)
	auth.POST("/setup-admin", handler.SetupAdmin)
}
