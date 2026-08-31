package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"xboard-go/pkg/logger"
)

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		cost := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.Duration("cost", cost),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if status >= 500 {
			logger.Get().Error("Request failed", fields...)
		} else if status >= 400 {
			logger.Get().Warn("Request warning", fields...)
		} else {
			logger.Get().Info("Request completed", fields...)
		}
	}
}
