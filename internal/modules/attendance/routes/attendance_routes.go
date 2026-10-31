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
	// Public routes: tanpa JWT
	publicAttendance := router.Group("/api/attendances")
	{
		publicAttendance.GET("/:id/check-in-photo", attendanceHandler.CheckInPhoto)
		publicAttendance.GET("/:id/check-out-photo", attendanceHandler.CheckOutPhoto)
	}

	// Protected routes: wajib JWT
	attendance := router.Group("/api/attendances")
	attendance.Use(middleware.AuthMiddleware())

	{
		attendance.POST("/check-in", attendanceHandler.CheckIn)
		attendance.POST("/check-out", attendanceHandler.CheckOut)

		attendance.GET("/paginated", attendanceHandler.FindAllPaginated)

		attendance.GET("/:id/check-in-detail", attendanceHandler.CheckInDetail)
		attendance.GET("/:id/check-out-detail", attendanceHandler.CheckOutDetail)

		attendance.GET("/today", attendanceHandler.GetTodayAttendance)
	}
}

// func RegisterAttendanceRoutes(
// 	router *gin.Engine,
// 	attendanceHandler *handler.AttendanceHandler,
// 	userRepo userRepository.UserRepository,
// 	roleRepo roleRepository.RoleRepository,
// ) {
// 	attendance := router.Group("/api/attendances")
// 	attendance.Use(middleware.AuthMiddleware())

// 	{
// 		attendance.POST("/check-in", attendanceHandler.CheckIn)
// 		// attendance.POST("/:id/check-out", attendanceHandler.CheckOut)

// 		attendance.POST("/check-out", attendanceHandler.CheckOut)

// 		attendance.GET("/:id/check-in-photo", attendanceHandler.CheckInPhoto)
// 		attendance.GET("/:id/check-out-photo", attendanceHandler.CheckOutPhoto)

// 		// Admin-only attendance endpoints
// 		// attendance.GET(
// 		// 	"/paginated",
// 		// 	middleware.AdminMiddleware(userRepo, roleRepo),
// 		// 	attendanceHandler.FindAllPaginated,
// 		// )

// 		attendance.GET("/paginated", attendanceHandler.FindAllPaginated)

// 		attendance.GET("/:id/check-in-detail", attendanceHandler.CheckInDetail)
// 		attendance.GET("/:id/check-out-detail", attendanceHandler.CheckOutDetail)

// 	}
// }
