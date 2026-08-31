package middleware

import (
	"context"
	"fmt"
	"time"

	"xboard-go/pkg/response"
	appredis "xboard-go/pkg/redis"

	"github.com/gin-gonic/gin"
)

const (
	loginAttemptPrefix = "login:attempt:"
	loginBlockPrefix   = "login:block:"
	registerPrefix     = "register:limit:"
)

// LoginSecurityMiddleware 登录安全中间件（防暴力破解）
func LoginSecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用密码限制
		enabled := getSettingBool("password_limit_enable", true)
		if !enabled {
			c.Next()
			return
		}

		client := appredis.Client()
		if client == nil {
			c.Next()
			return
		}

		ctx := context.Background()
		ip := c.ClientIP()

		// 检查是否被封禁
		blockKey := fmt.Sprintf("%s%s", loginBlockPrefix, ip)
		blocked, err := client.Exists(ctx, blockKey).Result()
		if err == nil && blocked > 0 {
			response.TooManyRequests(c, "登录尝试过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RecordLoginAttempt 记录登录失败尝试
func RecordLoginAttempt(ip string) {
	client := appredis.Client()
	if client == nil {
		return
	}

	ctx := context.Background()

	// 获取配置
	maxAttempts := 5
	blockDuration := 60 // 封禁时长（秒）

	// 尝试从设置读取
	if val := getSetting("password_limit_count", "5"); val != "" {
		fmt.Sscanf(val, "%d", &maxAttempts)
	}
	if val := getSetting("password_limit_expire", "60"); val != "" {
		fmt.Sscanf(val, "%d", &blockDuration)
	}

	attemptKey := fmt.Sprintf("%s%s", loginAttemptPrefix, ip)

	// 增加尝试次数
	count, err := client.Incr(ctx, attemptKey).Result()
	if err != nil {
		return
	}

	// 设置过期时间
	if count == 1 {
		client.Expire(ctx, attemptKey, time.Duration(blockDuration)*time.Second)
	}

	// 超过最大尝试次数，封禁 IP
	if int(count) >= maxAttempts {
		blockKey := fmt.Sprintf("%s%s", loginBlockPrefix, ip)
		client.Set(ctx, blockKey, 1, time.Duration(blockDuration)*time.Second)
		// 清除尝试计数
		client.Del(ctx, attemptKey)
	}
}

// ClearLoginAttempts 清除登录尝试记录（登录成功时调用）
func ClearLoginAttempts(ip string) {
	client := appredis.Client()
	if client == nil {
		return
	}

	ctx := context.Background()
	attemptKey := fmt.Sprintf("%s%s", loginAttemptPrefix, ip)
	blockKey := fmt.Sprintf("%s%s", loginBlockPrefix, ip)

	client.Del(ctx, attemptKey, blockKey)
}

// RegisterRateLimitMiddleware 注册频率限制中间件
func RegisterRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用注册限制
		enabled := getSettingBool("register_limit_by_ip_enable", false)
		if !enabled {
			c.Next()
			return
		}

		client := appredis.Client()
		if client == nil {
			c.Next()
			return
		}

		ctx := context.Background()
		ip := c.ClientIP()

		// 获取配置
		maxCount := 3
		windowSeconds := 3600 // 1小时

		if val := getSetting("register_limit_count", "3"); val != "" {
			fmt.Sscanf(val, "%d", &maxCount)
		}
		if val := getSetting("register_limit_expire", "3600"); val != "" {
			fmt.Sscanf(val, "%d", &windowSeconds)
		}

		key := fmt.Sprintf("%s%s", registerPrefix, ip)

		// 获取当前计数
		count, err := client.Get(ctx, key).Int()
		if err != nil && err.Error() != "redis: nil" {
			c.Next()
			return
		}

		if count >= maxCount {
			response.TooManyRequests(c, "注册过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RecordRegisterAttempt 记录注册尝试
func RecordRegisterAttempt(ip string) {
	client := appredis.Client()
	if client == nil {
		return
	}

	ctx := context.Background()

	// 获取配置
	windowSeconds := 3600
	if val := getSetting("register_limit_expire", "3600"); val != "" {
		fmt.Sscanf(val, "%d", &windowSeconds)
	}

	key := fmt.Sprintf("%s%s", registerPrefix, ip)

	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		return
	}

	if count == 1 {
		client.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)
	}
}

// EmailWhitelistCheck 邮箱白名单检查
func EmailWhitelistCheck(email string) (bool, string) {
	// 检查是否启用邮箱白名单
	enabled := getSettingBool("email_whitelist_enable", false)
	if !enabled {
		return true, ""
	}

	// 获取白名单后缀
	suffixes := getSetting("email_whitelist_suffix", "")
	if suffixes == "" {
		return true, ""
	}

	// 检查邮箱是否匹配白名单后缀
	for _, suffix := range splitAndTrim(suffixes, ",") {
		if suffix != "" && endsWith(email, suffix) {
			return true, ""
		}
	}

	return false, "该邮箱不在允许注册的白名单中"
}

// GmailLimitCheck Gmail 限制检查
func GmailLimitCheck(email string) (bool, string) {
	enabled := getSettingBool("email_gmail_limit_enable", false)
	if !enabled {
		return true, ""
	}

	if endsWith(email, "@gmail.com") {
		return false, "暂不支持 Gmail 邮箱注册"
	}

	return true, ""
}

// splitAndTrim 分割并去除空白
func splitAndTrim(s, sep string) []string {
	var result []string
	for _, part := range split(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// split 分割字符串
func split(s, sep string) []string {
	if s == "" {
		return nil
	}
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

// trimSpace 去除空白
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// endsWith 检查是否以指定后缀结尾
func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
