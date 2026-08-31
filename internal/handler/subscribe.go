package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
)

// SubscribeHandler 订阅处理器
type SubscribeHandler struct {
	subscribeService service.SubscribeService
}

// NewSubscribeHandler 创建订阅处理器
func NewSubscribeHandler(subscribeService service.SubscribeService) *SubscribeHandler {
	return &SubscribeHandler{
		subscribeService: subscribeService,
	}
}

// ClientSubscribe 客户端订阅接口
// @Summary 客户端订阅
// @Description 生成客户端订阅配置（支持Clash/V2Ray/Sing-box格式）
// @Tags 订阅
// @Produce plain
// @Param token query string true "订阅令牌"
// @Param format query string false "订阅格式 (clash/v2ray/sing-box)" default(clash)
// @Success 200 {string} string "订阅配置内容"
// @Router /api/v1/client/subscribe [get]
func (h *SubscribeHandler) ClientSubscribe(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.String(http.StatusBadRequest, "token is required")
		return
	}

	format := c.DefaultQuery("format", "clash")

	// 验证用户
	user, err := h.subscribeService.GetUserByToken(token)
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid subscribe token")
		return
	}

	// 检查用户状态
	if !user.CanUseService() {
		c.String(http.StatusForbidden, "Account is not available (banned/expired/no traffic)")
		return
	}

	// 生成订阅内容
	content, err := h.subscribeService.GenerateSubscribe(user, format)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate subscribe: "+err.Error())
		return
	}

	// 设置响应头
	c.Header("Content-Type", getContentType(format))
	c.Header("Content-Disposition", "attachment; filename=subscribe")
	c.Header("Profile-Update-Interval", "6") // Clash 更新间隔（小时）

	c.String(http.StatusOK, content)
}

// getContentType 根据格式返回 Content-Type
func getContentType(format string) string {
	switch format {
	case "clash", "clashmeta", "mihomo", "stash":
		return "application/yaml; charset=utf-8"
	case "v2ray", "v2":
		return "text/plain; charset=utf-8"
	case "sing-box", "singbox":
		return "application/json; charset=utf-8"
	case "surge", "surfboard", "loon":
		return "text/plain; charset=utf-8"
	case "quantumultx", "quanx", "qx":
		return "text/plain; charset=utf-8"
	default:
		return "text/plain; charset=utf-8"
	}
}
