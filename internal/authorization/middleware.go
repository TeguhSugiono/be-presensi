package authorization

import (
	"net/http"
	"strings"

	authMiddleware "be-presensi/internal/modules/auth/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type userRole struct {
	RoleName string `gorm:"column:role_name"`
}

func RequirePermission(db *gorm.DB, permission Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := authMiddleware.GetUserID(c)
		if !ok || userID <= 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid authenticated user",
			})
			return
		}

		var result userRole

		tx := db.Table("users").
			Select("roles.name AS role_name").
			Joins("JOIN roles ON roles.id = users.role_id").
			Where(
				"users.id = ? AND users.is_active = ? AND users.is_deleted = ? AND roles.is_deleted = ?",
				userID, true, false, false,
			).
			Scan(&result)

		if tx.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to verify user permissions",
			})
			return
		}

		if tx.RowsAffected == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "user is inactive, deleted, or not found",
			})
			return
		}

		if !HasPermission(
			strings.TrimSpace(result.RoleName),
			permission,
		) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "you do not have permission to access this resource",
			})
			return
		}

		c.Next()
	}
}
