package routes

import (
	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/daily_activity/handler"

	"github.com/gin-gonic/gin"
)

func RegisterDailyActivityRoutes(
	router *gin.Engine,
	dailyActivityHandler *handler.DailyActivityHandler,
) {
	group := router.Group("/api/daily-activities")
	group.Use(middleware.AuthMiddleware())
	{
		group.POST(
			"/attendances/:attendance_id",
			dailyActivityHandler.Create,
		)
		group.GET(
			"/attendances/:attendance_id",
			dailyActivityHandler.GetByAttendance,
		)
		group.PUT(
			"/attendances/:attendance_id",
			dailyActivityHandler.Update,
		)
	}
}
