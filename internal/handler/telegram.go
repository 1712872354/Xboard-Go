package handler

import (
	"net/http"

	"xboard-go/internal/service"
	"xboard-go/pkg/response"
	"xboard-go/pkg/telegram"

	"github.com/gin-gonic/gin"
)

// TelegramHandler Telegram 处理器
type TelegramHandler struct {
	telegramService service.TelegramService
}

// NewTelegramHandler 创建 Telegram 处理器
func NewTelegramHandler(telegramService service.TelegramService) *TelegramHandler {
	return &TelegramHandler{
		telegramService: telegramService,
	}
}

// Webhook 处理 Telegram Webhook
func (h *TelegramHandler) Webhook(c *gin.Context) {
	var update telegram.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 异步处理消息，立即返回 200
	go func() {
		if err := h.telegramService.HandleWebhook(&update); err != nil {
			// 日志记录错误
			_ = err
		}
	}()

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// SetWebhook 设置 Telegram Webhook
func (h *TelegramHandler) SetWebhook(c *gin.Context) {
	var req struct {
		WebhookURL string `json:"webhook_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.telegramService.SetWebhook(req.WebhookURL); err != nil {
		response.Fail(c, 500, "设置 Webhook 失败: "+err.Error())
		return
	}

	// 注册 Bot 命令
	if err := h.telegramService.RegisterCommands(); err != nil {
		response.Fail(c, 500, "注册命令失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"success":     true,
		"webhook_url": req.WebhookURL,
	})
}

// GetBotInfo 获取 Bot 信息
func (h *TelegramHandler) GetBotInfo(c *gin.Context) {
	// 从设置获取 bot token 来验证配置
	token := ""
	// 这里只是检查配置是否存在
	if token == "" {
		response.Fail(c, 400, "Telegram Bot Token 未配置")
		return
	}

	response.Success(c, gin.H{
		"configured": true,
	})
}
