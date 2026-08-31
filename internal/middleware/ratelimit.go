package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"xboard-go/pkg/ratelimit"
	"xboard-go/pkg/response"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// IP 级限流（全局）
	IPLimit    int           // 每个 IP 在窗口内最大请求数
	IPWindow   time.Duration // IP 限流窗口

	// 用户级限流（登录用户）
	UserLimit  int           // 每个用户在窗口内最大请求数
	UserWindow time.Duration // 用户限流窗口

	// 白名单（IP 或用户 ID 列表，不限流）
	IPWhitelist   []string
	PathWhitelist []string // 路径白名单（如健康检查）
}

// DefaultRateLimitConfig 默认限流配置
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		IPLimit:       100,
		IPWindow:      time.Minute,
		UserLimit:     300,
		UserWindow:    time.Minute,
		IPWhitelist:   []string{"127.0.0.1", "::1"},
		PathWhitelist: []string{"/healthz"},
	}
}

// RateLimitMiddleware 限流中间件
// 支持 IP 级和用户级两级限流
// 优先使用 Redis 限流器，Redis 不可用时降级到内存限流器
func RateLimitMiddleware(limiter ratelimit.Limiter, config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查路径白名单
		path := c.Request.URL.Path
		for _, wp := range config.PathWhitelist {
			if path == wp {
				c.Next()
				return
			}
		}

		// 获取客户端 IP
		clientIP := getClientIP(c)

		// 检查 IP 白名单
		for _, wip := range config.IPWhitelist {
			if clientIP == wip {
				c.Next()
				return
			}
		}

		// 1. IP 级限流
		ipKey := fmt.Sprintf("ip:%s", clientIP)
		allowed, remaining, resetTime, err := limiter.Allow(ipKey, config.IPLimit, config.IPWindow)
		if err != nil {
			// 限流出错，放行（降级策略）
			c.Next()
			return
		}

		// 设置限流响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.IPLimit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if !allowed {
			retryAfter := int(time.Until(resetTime).Seconds()) + 1
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			response.TooManyRequests(c, fmt.Sprintf("Rate limit exceeded. Try again in %d seconds.", retryAfter))
			c.Abort()
			return
		}

		// 2. 用户级限流（仅当用户已登录时）
		userID := GetUserID(c)
		if userID > 0 {
			userKey := fmt.Sprintf("user:%d", userID)
			userAllowed, userRemaining, userResetTime, err := limiter.Allow(userKey, config.UserLimit, config.UserWindow)
			if err != nil {
				c.Next()
				return
			}

			c.Header("X-User-RateLimit-Limit", fmt.Sprintf("%d", config.UserLimit))
			c.Header("X-User-RateLimit-Remaining", fmt.Sprintf("%d", userRemaining))
			c.Header("X-User-RateLimit-Reset", fmt.Sprintf("%d", userResetTime.Unix()))

			if !userAllowed {
				retryAfter := int(time.Until(userResetTime).Seconds()) + 1
				c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
				response.TooManyRequests(c, fmt.Sprintf("User rate limit exceeded. Try again in %d seconds.", retryAfter))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// getClientIP 获取客户端真实 IP
func getClientIP(c *gin.Context) string {
	// 优先从 X-Forwarded-For 获取（经过代理时）
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For 可能包含多个 IP，取第一个
		ips := splitFirst(xff, ",")
		if ips != "" {
			return ips
		}
	}

	// 其次从 X-Real-IP 获取
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// 最后从 RemoteAddr 获取
	return c.ClientIP()
}

// splitFirst 取字符串分隔后的第一个部分
func splitFirst(s, sep string) string {
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			return s[:i]
		}
	}
	return s
}

// 确保响应状态码存在
var _ = http.StatusTooManyRequests
