package middleware

import (
	"github.com/gin-gonic/gin"
	"xboard-go/pkg/response"
)

// RBAC 角色权限中间件
func RBAC(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)
		if userRole == "" {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		for _, role := range roles {
			if role == userRole {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "Insufficient permissions")
		c.Abort()
	}
}

// AdminRequired 管理员权限中间件
func AdminRequired() gin.HandlerFunc {
	return RBAC("admin")
}

// UserRequired 用户权限中间件（已登录用户即可）
func UserRequired() gin.HandlerFunc {
	return RBAC("user", "admin")
}
