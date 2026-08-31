package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"xboard-go/pkg/jwt"
	"xboard-go/pkg/response"
)

// Context keys
const (
	UserIDKey   = "user_id"
	UserRoleKey = "user_role"
	UserEmailKey = "user_email"
)

// JWTAuth JWT 认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization format")
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// 将用户信息存入 context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserRoleKey, claims.Role)
		c.Set(UserEmailKey, claims.Email)

		c.Next()
	}
}

// GetUserID 从 context 获取用户ID
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0
	}
	return userID.(uint)
}

// GetUserRole 从 context 获取用户角色
func GetUserRole(c *gin.Context) string {
	role, exists := c.Get(UserRoleKey)
	if !exists {
		return ""
	}
	return role.(string)
}

// IsAdmin 检查当前用户是否是管理员
func IsAdmin(c *gin.Context) bool {
	role := GetUserRole(c)
	return role == "admin"
}
