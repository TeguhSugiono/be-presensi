package routes

import (
	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/school_location/handler"

	"github.com/gin-gonic/gin"
)

func RegisterSchoolLocationRoutes(
	router *gin.RouterGroup,
	handler *handler.SchoolLocationHandler,
) {
	locations := router.Group("/school-locations")

	locations.Use(middleware.AuthMiddleware())

	locations.GET("", handler.FindAll)
	locations.GET("/paginated", handler.FindAllPaginated)
	locations.GET("/:id", handler.FindByID)
	locations.POST("", handler.Create)
	locations.PUT("/:id", handler.Update)
	locations.DELETE("/:id", handler.Delete)
}
