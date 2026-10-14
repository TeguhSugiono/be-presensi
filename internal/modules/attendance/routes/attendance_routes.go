package routes

import (
	"be-presensi/internal/modules/attendance/handler"
	"be-presensi/internal/modules/auth/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAttendanceRoutes(
	router *gin.Engine,
	attendanceHandler *handler.AttendanceHandler,
) {
	attendance := router.Group("/api/attendances")
	attendance.Use(middleware.AuthMiddleware())
	{
		attendance.POST("/check-in", attendanceHandler.CheckIn)
		attendance.POST("/:id/check-out", attendanceHandler.CheckOut)
	}
}
