package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"xboard-go/pkg/logger"
	"xboard-go/pkg/response"
)

// Recovery 错误恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Sugar().Errorf("Panic recovered: %v\nStack: %s", err, debug.Stack())
				response.InternalError(c, "Internal Server Error")
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
