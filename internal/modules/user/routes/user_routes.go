package routes

import (
	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/user/handler"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(router *gin.RouterGroup, handler *handler.UserHandler) {
	users := router.Group("/users")

	users.Use(middleware.AuthMiddleware())

	users.GET("", handler.FindAll)
	users.GET("/:id", handler.FindByID)
	users.POST("", handler.Create)
	users.PUT("/:id", handler.Update)
	users.DELETE("/:id", handler.Delete)
	users.GET("/paginated", handler.FindAllPaginated)
	users.POST("/:id/change-password", handler.ChangePassword)
}
