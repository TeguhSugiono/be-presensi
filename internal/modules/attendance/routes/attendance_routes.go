package routes

import (
	"be-presensi/internal/modules/attendance/handler"
	"be-presensi/internal/modules/auth/middleware"
	roleRepository "be-presensi/internal/modules/role/repository"
	userRepository "be-presensi/internal/modules/user/repository"

	"github.com/gin-gonic/gin"
)

func RegisterAttendanceRoutes(
	router *gin.Engine,
	attendanceHandler *handler.AttendanceHandler,
	userRepo userRepository.UserRepository,
	roleRepo roleRepository.RoleRepository,
) {
	attendance := router.Group("/api/attendances")
	attendance.Use(middleware.AuthMiddleware())

	{
		attendance.POST("/check-in", attendanceHandler.CheckIn)
		attendance.POST("/:id/check-out", attendanceHandler.CheckOut)

		attendance.GET("/:id/check-in-photo", attendanceHandler.CheckInPhoto)
		attendance.GET("/:id/check-out-photo", attendanceHandler.CheckOutPhoto)

		// Admin-only attendance endpoints
		attendance.GET(
			"/paginated",
			middleware.AdminMiddleware(userRepo, roleRepo),
			attendanceHandler.FindAllPaginated,
		)

	}
}
