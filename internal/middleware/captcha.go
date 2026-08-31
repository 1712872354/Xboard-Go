package middleware

import (
	"fmt"

	"xboard-go/internal/model"
	"xboard-go/pkg/captcha"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// CaptchaMiddleware 验证码中间件
func CaptchaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从设置读取验证码配置
		enabled := getSettingBool("captcha_enable", false)
		if !enabled {
			c.Next()
			return
		}

		provider := getSetting("captcha_type", "turnstile")
		siteKey := getSetting("captcha_site_key", "")
		secretKey := getSetting("captcha_secret_key", "")
		minScore := getSettingFloat("captcha_min_score", 0.5)

		if secretKey == "" {
			c.Next()
			return
		}

		cfg := captcha.Config{
			Provider:  captcha.Provider(provider),
			SiteKey:   siteKey,
			SecretKey: secretKey,
			MinScore:  minScore,
		}

		svc := captcha.NewCaptchaService(cfg)

		// 从请求中获取 token
		token := c.Query("captcha_token")
		if token == "" {
			token = c.PostForm("captcha_token")
		}
		if token == "" {
			// 尝试从 header 获取
			token = c.GetHeader("X-Captcha-Token")
		}

		if token == "" {
			response.BadRequest(c, "验证码 token 缺失")
			c.Abort()
			return
		}

		remoteIP := c.ClientIP()
		ok, err := svc.Verify(token, remoteIP)
		if err != nil {
			response.Fail(c, 500, "验证码验证失败: "+err.Error())
			c.Abort()
			return
		}

		if !ok {
			response.Fail(c, 400, "验证码验证失败")
			c.Abort()
			return
		}

		c.Next()
	}
}

// getSetting 从数据库获取设置值
func getSetting(key, defaultValue string) string {
	db := database.Get()
	if db == nil {
		return defaultValue
	}
	var setting model.Setting
	if err := db.Where("`key` = ?", key).First(&setting).Error; err != nil {
		return defaultValue
	}
	if setting.Value == "" {
		return defaultValue
	}
	return setting.Value
}

// getSettingBool 获取布尔类型设置
func getSettingBool(key string, defaultValue bool) bool {
	val := getSetting(key, "")
	if val == "" {
		return defaultValue
	}
	return val == "1" || val == "true" || val == "yes"
}

// getSettingFloat 获取浮点类型设置
func getSettingFloat(key string, defaultValue float64) float64 {
	val := getSetting(key, "")
	if val == "" {
		return defaultValue
	}
	var result float64
	if _, err := fmt.Sscanf(val, "%f", &result); err != nil {
		return defaultValue
	}
	return result
}
