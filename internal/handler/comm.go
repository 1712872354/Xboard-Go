package handler

import (
	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// CommHandler 公共处理器
type CommHandler struct{}

// NewCommHandler 创建公共处理器
func NewCommHandler() *CommHandler {
	return &CommHandler{}
}

// GetConfig 获取全局配置（用户端）
func (h *CommHandler) GetConfig(c *gin.Context) {
	db := database.Get()

	keys := []string{
		"app_name", "app_url", "app_description", "logo", "tos_url",
		"currency", "currency_symbol", "stop_register",
		"subscribe_path", "telegram_discuss_link",
		"windows_version", "windows_download_url",
		"macos_version", "macos_download_url",
		"android_version", "android_download_url",
	}

	config := make(map[string]string)
	for _, key := range keys {
		var setting model.Setting
		if err := db.Where("`key` = ?", key).First(&setting).Error; err == nil {
			config[key] = setting.Value
		}
	}

	response.Success(c, config)
}

// GetStripePublicKey 获取 Stripe 公钥
func (h *CommHandler) GetStripePublicKey(c *gin.Context) {
	db := database.Get()

	var setting model.Setting
	if err := db.Where("`key` = ?", "stripe_public_key").First(&setting).Error; err != nil {
		response.Success(c, gin.H{"public_key": ""})
		return
	}

	response.Success(c, gin.H{"public_key": setting.Value})
}
