package routes

import (
	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/work_schedule/handler"

	"github.com/gin-gonic/gin"
)

func RegisterWorkScheduleRoutes(router *gin.RouterGroup, handler *handler.WorkScheduleHandler) {
	workSchedules := router.Group("/work-schedules")

	workSchedules.Use(middleware.AuthMiddleware())

	workSchedules.GET("", handler.FindAll)
	workSchedules.GET("/:id", handler.FindByID)
	workSchedules.POST("", handler.Create)
	workSchedules.PUT("/:id", handler.Update)
	workSchedules.DELETE("/:id", handler.Delete)
	workSchedules.GET("/paginated", handler.FindAllPaginated)
}
