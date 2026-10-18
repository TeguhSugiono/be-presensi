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

	schoolLocationHandler "be-presensi/internal/modules/school_location/handler"
	schoolLocationRepository "be-presensi/internal/modules/school_location/repository"
	schoolLocationRoutes "be-presensi/internal/modules/school_location/routes"
	schoolLocationService "be-presensi/internal/modules/school_location/service"

	attendanceHandler "be-presensi/internal/modules/attendance/handler"
	attendanceRepository "be-presensi/internal/modules/attendance/repository"
	attendanceRoutes "be-presensi/internal/modules/attendance/routes"
	attendanceService "be-presensi/internal/modules/attendance/service"

	dailyActivityRepository "be-presensi/internal/modules/daily_activity/repository"

	dailyActivityHandler "be-presensi/internal/modules/daily_activity/handler"
	dailyActivityRoutes "be-presensi/internal/modules/daily_activity/routes"
	dailyActivityService "be-presensi/internal/modules/daily_activity/service"

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
	//roleRoutes.RegisterRoleRoutes(api, roleHdl, db)

	// =========================
	// Work Schedule Module
	// =========================
	workScheduleRepo := workScheduleRepository.NewWorkScheduleRepository(db)
	workScheduleSvc := workScheduleService.NewWorkScheduleService(workScheduleRepo)
	workScheduleHdl := workScheduleHandler.NewWorkScheduleHandler(workScheduleSvc)

	workScheduleRoutes.RegisterWorkScheduleRoutes(api, workScheduleHdl)

	// =========================
	// School Location Module
	// =========================
	schoolLocationRepo := schoolLocationRepository.NewSchoolLocationRepository(db)
	schoolLocationSvc := schoolLocationService.NewSchoolLocationService(schoolLocationRepo)
	schoolLocationHdl := schoolLocationHandler.NewSchoolLocationHandler(schoolLocationSvc)

	schoolLocationRoutes.RegisterSchoolLocationRoutes(api, schoolLocationHdl)

	// =========================
	// User Module
	// =========================
	userRepo := userRepository.NewUserRepository(db)

	// =========================
	// Attendance Module
	// =========================
	attendanceRepo := attendanceRepository.NewAttendanceRepository(db)

	dailyActivityRepo := dailyActivityRepository.NewDailyActivityRepository(db)

	attendanceSvc := attendanceService.NewAttendanceService(
		db,
		attendanceRepo,
		dailyActivityRepo,
		userRepo,
		roleRepo,
		workScheduleRepo,
		schoolLocationRepo,
	)

	attendanceHdl := attendanceHandler.NewAttendanceHandler(attendanceSvc)

	attendanceRoutes.RegisterAttendanceRoutes(
		router,
		attendanceHdl,
		userRepo,
		roleRepo,
	)

	// =========================
	// Daily Activity Module
	// =========================
	dailyActivitySvc := dailyActivityService.NewDailyActivityService(
		db,
		attendanceRepo,
		dailyActivityRepo,
	)

	dailyActivityHdl := dailyActivityHandler.NewDailyActivityHandler(
		dailyActivitySvc,
	)

	dailyActivityRoutes.RegisterDailyActivityRoutes(
		router,
		dailyActivityHdl,
		userRepo,
		roleRepo,
	)

	userSvc := userService.NewUserService(userRepo, db)
	userHdl := userHandler.NewUserHandler(userSvc)

	// =========================
	// Auth Module
	// =========================
	//authSvc := authService.NewAuthService(userRepo)
	// authSvc := authService.NewAuthService(userRepo, db)

	authSvc := authService.NewAuthService(
		userRepo,
		roleRepo,
		db,
	)

	authHdl := authHandler.NewAuthHandler(authSvc)

	authRoutes.RegisterAuthRoutes(api, authHdl)

	userRoutes.RegisterUserRoutes(api, userHdl)

	address := ":" + cfg.AppPort

	router.Run(address)
}
