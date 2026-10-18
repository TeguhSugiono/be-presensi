package middleware

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	roleRepository "be-presensi/internal/modules/role/repository"
	userRepository "be-presensi/internal/modules/user/repository"

	"be-presensi/internal/modules/auth/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "authorization header is required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid authorization header",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		secret := os.Getenv("JWT_SECRET")

		if secret == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "JWT_SECRET is not configured",
			})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(
			tokenString,
			&service.JWTClaims{},
			func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, jwt.ErrSignatureInvalid
				}

				return []byte(secret), nil
			},
		)

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*service.JWTClaims)

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid token claims",
			})
			c.Abort()
			return
		}

		userID := claims.UserID

		if userID <= 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid user id in token",
			})
			c.Abort()
			return
		}

		c.Set("user_id", userID)

		c.Next()
	}
}

func GetUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get("user_id")

	if !exists {
		return 0, false
	}

	userID, ok := value.(int64)

	if ok {
		return userID, true
	}

	// Fallback jika suatu saat context berisi string.
	if valueString, ok := value.(string); ok {
		parsed, err := strconv.ParseInt(valueString, 10, 64)

		if err != nil {
			return 0, false
		}

		return parsed, true
	}

	return 0, false
}

func AdminMiddleware(
	userRepo userRepository.UserRepository,
	roleRepo roleRepository.RoleRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok || userID <= 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid authenticated user",
			})
			return
		}

		user, err := userRepo.FindByID(userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "user not found",
				})
				return
			}

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to verify user access",
			})
			return
		}

		if user == nil || !user.IsActive || user.IsDeleted {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "active user account required",
			})
			return
		}

		role, err := roleRepo.FindByID(c.Request.Context(), user.RoleID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "admin access required",
				})
				return
			}

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to verify user role",
			})
			return
		}

		if role == nil || role.IsDeleted || !strings.EqualFold(role.Name, "Admin") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "admin access required",
			})
			return
		}

		c.Next()
	}
}
