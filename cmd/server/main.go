package main

import (
	"net/http"

	"be-presensi/internal/config"
	"be-presensi/internal/database"

	roleHandler "be-presensi/internal/modules/role/handler"
	roleRepository "be-presensi/internal/modules/role/repository"
	roleRoutes "be-presensi/internal/modules/role/routes"
	roleService "be-presensi/internal/modules/role/service"

	workScheduleHandler "be-presensi/internal/modules/work_schedule/handler"
	workScheduleRepository "be-presensi/internal/modules/work_schedule/repository"
	workScheduleRoutes "be-presensi/internal/modules/work_schedule/routes"
	workScheduleService "be-presensi/internal/modules/work_schedule/service"

	userHandler "be-presensi/internal/modules/user/handler"
	userRepository "be-presensi/internal/modules/user/repository"
	userRoutes "be-presensi/internal/modules/user/routes"
	userService "be-presensi/internal/modules/user/service"

	authHandler "be-presensi/internal/modules/auth/handler"
	authRoutes "be-presensi/internal/modules/auth/routes"
	authService "be-presensi/internal/modules/auth/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg)

	router := gin.Default()

	// Health Check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Backend Presensi is running",
		})
	})

	api := router.Group("/api")

	// =========================
	// Role Module
	// =========================
	roleRepo := roleRepository.NewRoleRepository(db)
	roleSvc := roleService.NewRoleService(roleRepo)
	roleHdl := roleHandler.NewRoleHandler(roleSvc)

	roleRoutes.RegisterRoleRoutes(api, roleHdl)

	// =========================
	// Work Schedule Module
	// =========================
	workScheduleRepo := workScheduleRepository.NewWorkScheduleRepository(db)
	workScheduleSvc := workScheduleService.NewWorkScheduleService(workScheduleRepo)
	workScheduleHdl := workScheduleHandler.NewWorkScheduleHandler(workScheduleSvc)

	workScheduleRoutes.RegisterWorkScheduleRoutes(api, workScheduleHdl)

	// =========================
	// User Module
	// =========================
	userRepo := userRepository.NewUserRepository(db)
	userSvc := userService.NewUserService(userRepo, db)
	userHdl := userHandler.NewUserHandler(userSvc)

	// =========================
	// Auth Module
	// =========================
	//authSvc := authService.NewAuthService(userRepo)
	authSvc := authService.NewAuthService(userRepo, db)
	authHdl := authHandler.NewAuthHandler(authSvc)

	authRoutes.RegisterAuthRoutes(api, authHdl)

	userRoutes.RegisterUserRoutes(api, userHdl)

	address := ":" + cfg.AppPort

	router.Run(address)
}
