package routes

import (
	"be-presensi/internal/modules/activity_comment/handler"
	"be-presensi/internal/modules/auth/middleware"
	roleRepository "be-presensi/internal/modules/role/repository"
	userRepository "be-presensi/internal/modules/user/repository"

	"github.com/gin-gonic/gin"
)

func RegisterActivityCommentRoutes(
	router *gin.Engine,
	activityCommentHandler *handler.ActivityCommentHandler,
	userRepo userRepository.UserRepository,
	roleRepo roleRepository.RoleRepository,
) {
	group := router.Group("/api/daily-activities")
	group.Use(middleware.AuthMiddleware())

	group.POST(
		"/activity-comments",
		middleware.AdminMiddleware(userRepo, roleRepo),
		activityCommentHandler.Create,
	)

	group.GET(
		"/activity-comments/daily-activities/:id",
		activityCommentHandler.GetByDailyActivityID,
	)
}
