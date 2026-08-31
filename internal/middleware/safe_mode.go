package middleware

import (
	"net/http"
	"net/url"

	"xboard-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// SafeModeMiddleware 安全模式中间件
// 验证请求 Host 是否匹配配置的 app_url
func SafeModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用安全模式
		enabled := getSettingBool("safe_mode_enable", false)
		if !enabled {
			c.Next()
			return
		}

		// 获取配置的 app_url
		appURL := getSetting("app_url", "")
		if appURL == "" {
			c.Next()
			return
		}

		// 解析配置的域名
		parsedURL, err := url.Parse(appURL)
		if err != nil {
			c.Next()
			return
		}

		configHost := parsedURL.Hostname()
		if configHost == "" {
			c.Next()
			return
		}

		// 获取请求的 Host
		requestHost := c.Request.Host
		if requestHost == "" {
			requestHost = c.GetHeader("Host")
		}

		// 提取主机名（移除端口）
		if idx := indexOf(requestHost, ':'); idx > 0 {
			requestHost = requestHost[:idx]
		}

		// 验证域名匹配
		if requestHost != configHost {
			c.JSON(http.StatusForbidden, response.Response{
				Code:    403,
				Message: "Access denied: invalid host",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// indexOf 查找字符在字符串中的位置
func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
