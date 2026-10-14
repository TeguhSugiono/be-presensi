package routes

import (
	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/modules/role/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(
	router *gin.RouterGroup,
	handler *handler.RoleHandler,
) {
	roles := router.Group("/roles")

	roles.Use(middleware.AuthMiddleware())

	roles.GET("", handler.FindAll)
	roles.GET("/:id", handler.FindByID)
	roles.POST("", handler.Create)
	roles.PUT("/:id", handler.Update)
	roles.DELETE("/:id", handler.Delete)
	roles.GET("/paginated", handler.FindAllPaginated)
}

// package routes

// import (
// 	"be-presensi/internal/authorization"
// 	authMiddleware "be-presensi/internal/modules/auth/middleware"
// 	"be-presensi/internal/modules/role/handler"

// 	"github.com/gin-gonic/gin"
// 	"gorm.io/gorm"
// )

// func RegisterRoleRoutes(
// 	router *gin.RouterGroup,
// 	handler *handler.RoleHandler,
// 	db *gorm.DB,
// ) {
// 	roles := router.Group("/roles")

// 	roles.Use(authMiddleware.AuthMiddleware())

// 	roles.GET(
// 		"",
// 		authorization.RequirePermission(db, authorization.RolesView),
// 		handler.FindAll,
// 	)
// 	roles.GET(
// 		"/paginated",
// 		authorization.RequirePermission(db, authorization.RolesView),
// 		handler.FindAllPaginated,
// 	)
// 	roles.GET(
// 		"/:id",
// 		authorization.RequirePermission(db, authorization.RolesView),
// 		handler.FindByID,
// 	)
// 	roles.POST(
// 		"",
// 		authorization.RequirePermission(db, authorization.RolesCreate),
// 		handler.Create,
// 	)
// 	roles.PUT(
// 		"/:id",
// 		authorization.RequirePermission(db, authorization.RolesUpdate),
// 		handler.Update,
// 	)
// 	roles.DELETE(
// 		"/:id",
// 		authorization.RequirePermission(db, authorization.RolesDelete),
// 		handler.Delete,
// 	)
// }
