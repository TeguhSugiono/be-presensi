package routes

import (
	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/role/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(
	router *gin.RouterGroup,
	handler *handler.RoleHandler,
) {
	roles := router.Group("/roles")

	roles.Use(middleware.AuthMiddleware())

	roles.GET("", handler.FindAll)
	roles.GET("/:id", handler.FindByID)
	roles.POST("", handler.Create)
	roles.PUT("/:id", handler.Update)
	roles.DELETE("/:id", handler.Delete)
}
