package handler

import (
	"github.com/gin-gonic/gin"
	"xboard-go/internal/model"
	"xboard-go/internal/service"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"
)

// SettingHandler 系统设置处理器
type SettingHandler struct {
	settingService service.SettingService
}

// NewSettingHandler 创建系统设置处理器
func NewSettingHandler(settingService service.SettingService) *SettingHandler {
	return &SettingHandler{
		settingService: settingService,
	}
}

// GetSettings 获取所有设置（管理员）
func (h *SettingHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingService.GetAll()
	if err != nil {
		response.InternalError(c, "获取设置失败")
		return
	}
	response.Success(c, settings)
}

// GetSettingsByGroup 获取分组设置（管理员）
func (h *SettingHandler) GetSettingsByGroup(c *gin.Context) {
	group := c.Param("group")
	if group == "" {
		response.BadRequest(c, "分组名称不能为空")
		return
	}

	settings, err := h.settingService.GetByGroup(group)
	if err != nil {
		response.InternalError(c, "获取设置失败")
		return
	}
	response.Success(c, settings)
}

// UpdateSettingsRequest 更新设置请求
type UpdateSettingsRequest struct {
	Settings []SettingItem `json:"settings" binding:"required,min=1"`
}

// SettingItem 设置项
type SettingItem struct {
	Key    string `json:"key" binding:"required"`
	Value  string `json:"value"`
	Group  string `json:"group"`
	Remark string `json:"remark"`
}

// UpdateSettings 批量更新设置（管理员）
func (h *SettingHandler) UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	settings := make([]model.Setting, len(req.Settings))
	for i, item := range req.Settings {
		settings[i] = model.Setting{
			Key:    item.Key,
			Value:  item.Value,
			Group:  item.Group,
			Remark: item.Remark,
		}
	}

	if err := h.settingService.SetBatch(settings); err != nil {
		response.InternalError(c, "更新设置失败")
		return
	}

	response.Success(c, nil)
}

// UpdateSettingRequest 更新单个设置请求
type UpdateSettingRequest struct {
	Value  string `json:"value"`
	Group  string `json:"group"`
	Remark string `json:"remark"`
}

// UpdateSetting 更新单个设置（管理员）
func (h *SettingHandler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "设置键名不能为空")
		return
	}

	var req UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.settingService.Set(key, req.Value, req.Group, req.Remark); err != nil {
		response.InternalError(c, "更新设置失败")
		return
	}

	response.Success(c, nil)
}

// DeleteSetting 删除设置（管理员）
func (h *SettingHandler) DeleteSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "设置键名不能为空")
		return
	}

	if err := h.settingService.Delete(key); err != nil {
		response.InternalError(c, "删除设置失败")
		return
	}

	response.Success(c, nil)
}

// GetPublicSettings 获取公开设置（无需认证）
func (h *SettingHandler) GetPublicSettings(c *gin.Context) {
	publicKeys := []string{
		model.SettingKeyAppName,
		model.SettingKeyAppURL,
		model.SettingKeyAppDescription,
		model.SettingKeyAppLogo,
		model.SettingKeyTOSURL,
		model.SettingKeyThemeFrontend,
		"captcha_enable",
		"captcha_type",
		"captcha_site_key",
		"stop_register",
		"email_verify",
		"tos_url",
		"telegram_discuss_link",
		"windows_version",
		"windows_download_url",
		"macos_version",
		"macos_download_url",
		"android_version",
		"android_download_url",
	}

	settings := make(map[string]string)
	for _, key := range publicKeys {
		value, err := h.settingService.GetByKey(key)
		if err != nil {
			continue
		}
		settings[key] = value
	}

	response.Success(c, settings)
}

// TestSendEmail 测试发送邮件（管理员）
func (h *SettingHandler) TestSendEmail(c *gin.Context) {
	// 获取当前用户邮箱
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	// 获取用户邮箱
	var user model.User
	if err := database.Get().First(&user, userID).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	mailService := service.NewMailService()
	if err := mailService.SendEmail(user.Email, "XBoard 测试邮件", "notify", map[string]interface{}{
		"name":    "XBoard",
		"content": "这是一封测试邮件，如果您收到此邮件，说明邮件配置正确。",
	}); err != nil {
		response.Fail(c, 500, "发送失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "测试邮件已发送至 " + user.Email,
	})
}

// GetConfigMappings 获取配置映射（分组）
func (h *SettingHandler) GetConfigMappings(c *gin.Context) {
	groups := map[string][]string{
		"site": {
			"app_name", "app_url", "app_description", "logo", "tos_url",
			"currency", "currency_symbol", "stop_register", "try_out_plan_id", "try_out_hour",
		},
		"subscribe": {
			"subscribe_path", "subscribe_url", "plan_change_enable", "surplus_enable",
			"reset_traffic_method", "show_info_to_server_enable", "show_protocol_to_server_enable",
			"default_remind_expire", "default_remind_traffic",
		},
		"server": {
			"server_token", "server_pull_interval", "server_push_interval", "device_limit_mode",
		},
		"email": {
			"email_host", "email_port", "email_username", "email_password",
			"email_encryption", "email_from_address", "remind_mail_enable",
		},
		"telegram": {
			"telegram_bot_enable", "telegram_bot_token", "telegram_webhook_url", "telegram_discuss_link",
		},
		"safe": {
			"email_verify", "safe_mode_enable", "secure_path",
			"email_whitelist_enable", "email_whitelist_suffix", "email_gmail_limit_enable",
			"captcha_enable", "captcha_type", "captcha_site_key", "captcha_secret_key",
			"captcha_min_score",
			"register_limit_by_ip_enable", "register_limit_count", "register_limit_expire",
			"password_limit_enable", "password_limit_count", "password_limit_expire",
		},
		"invite": {
			"invite_force", "invite_commission", "invite_gen_limit", "invite_never_expire",
			"commission_first_time_enable", "commission_auto_check_enable",
			"commission_withdraw_limit", "commission_withdraw_method",
			"withdraw_close_enable",
		},
		"frontend": {
			"frontend_theme", "frontend_theme_sidebar", "frontend_theme_header",
			"frontend_theme_color", "frontend_background_url",
		},
		"app": {
			"windows_version", "windows_download_url",
			"macos_version", "macos_download_url",
			"android_version", "android_download_url",
		},
	}

	result := make(map[string]map[string]string)
	for group, keys := range groups {
		result[group] = make(map[string]string)
		for _, key := range keys {
			value, err := h.settingService.GetByKey(key)
			if err != nil {
				result[group][key] = ""
			} else {
				result[group][key] = value
			}
		}
	}

	response.Success(c, result)
}
