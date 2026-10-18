package routes

import (
	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/daily_activity/handler"
	roleRepository "be-presensi/internal/modules/role/repository"
	userRepository "be-presensi/internal/modules/user/repository"

	"github.com/gin-gonic/gin"
)

func RegisterDailyActivityRoutes(
	router *gin.Engine,
	dailyActivityHandler *handler.DailyActivityHandler,
	userRepo userRepository.UserRepository,
	roleRepo roleRepository.RoleRepository,
) {
	group := router.Group("/api/daily-activities")
	group.Use(middleware.AuthMiddleware())

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

	group.GET(
		"/by-date",
		dailyActivityHandler.GetByDate,
	)

	group.POST(
		"/attendances/:attendance_id/upsert",
		dailyActivityHandler.Upsert,
	)

	group.GET(
		"/paginated",
		middleware.AdminMiddleware(userRepo, roleRepo),
		dailyActivityHandler.FindAllPaginated,
	)
}
