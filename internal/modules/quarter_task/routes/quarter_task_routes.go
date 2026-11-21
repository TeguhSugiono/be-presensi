package routes

import (
	"be-presensi/internal/modules/auth/middleware"
	quarterTaskHandler "be-presensi/internal/modules/quarter_task/handler"
	quarterTaskDetailHandler "be-presensi/internal/modules/quarter_task_detail/handler"
	roleRepository "be-presensi/internal/modules/role/repository"
	userRepository "be-presensi/internal/modules/user/repository"

	"github.com/gin-gonic/gin"
)

func RegisterQuarterTaskRoutes(
	router *gin.Engine,
	quarterTaskHandler *quarterTaskHandler.QuarterTaskHandler,
	quarterTaskDetailHandler *quarterTaskDetailHandler.QuarterTaskDetailHandler,
	userRepo userRepository.UserRepository,
	roleRepo roleRepository.RoleRepository,
) {
	// Semua endpoint Quarter Task wajib login.
	group := router.Group("/api/quarter-tasks")
	group.Use(middleware.AuthMiddleware())

	// ==========================================
	// ADMIN ROUTES
	// ==========================================
	admin := group.Group("")
	admin.Use(middleware.AdminMiddleware(userRepo, roleRepo))

	admin.POST(
		"",
		quarterTaskHandler.Create,
	)

	admin.PUT(
		"/:id",
		quarterTaskHandler.Update,
	)

	admin.DELETE(
		"/:id",
		quarterTaskHandler.Delete,
	)

	admin.GET(
		"",
		quarterTaskHandler.FindAllPaginated,
	)

	admin.GET(
		"/:id/submissions",
		quarterTaskHandler.GetSubmissions,
	)

	admin.GET(
		"/:id",
		quarterTaskHandler.GetByID,
	)

	// ==========================================
	// GURU / KARYAWAN ROUTES
	// ==========================================
	userGroup := router.Group("/api/user/quarter-tasks")
	userGroup.Use(middleware.AuthMiddleware())

	userGroup.GET(
		"",
		quarterTaskDetailHandler.FindAllTasks,
	)

	userGroup.GET(
		"/:id",
		quarterTaskDetailHandler.GetByTask,
	)

	userGroup.POST(
		"/:id",
		quarterTaskDetailHandler.Upsert,
	)
}
