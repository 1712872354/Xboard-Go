package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"xboard-go/config"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			// 检查Origin是否在白名单中
			if isOriginAllowed(origin) {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Max-Age", "86400")
			}
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// isOriginAllowed 检查Origin是否在允许列表中
func isOriginAllowed(origin string) bool {
	cfg := config.Get()
	allowedOrigins := cfg.Server.AllowedOrigins
	
	// 如果没有配置允许的源，则拒绝所有跨域请求
	if len(allowedOrigins) == 0 {
		return false
	}
	
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	
	return false
}
