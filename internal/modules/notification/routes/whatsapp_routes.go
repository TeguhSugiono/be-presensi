package routes

import (
	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/notification/handler"

	"github.com/gin-gonic/gin"
)

func RegisterWhatsAppRoutes(
	router *gin.Engine,
	whatsappHandler *handler.WhatsAppHandler,
) {
	whatsapp := router.Group("/api/notifications/whatsapp")
	whatsapp.Use(middleware.AuthMiddleware())

	{
		whatsapp.POST("/test", whatsappHandler.TestSend)
	}
}
