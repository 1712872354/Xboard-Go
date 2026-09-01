package handler

import (
	"github.com/gin-gonic/gin"
	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"
	"xboard-go/pkg/utils"
)

// TemplateHandler 模板处理器
type TemplateHandler struct{}

// NewTemplateHandler 创建模板处理器
func NewTemplateHandler() *TemplateHandler {
	return &TemplateHandler{}
}

// GetEmailTemplate 获取邮件模板
// @Summary 获取邮件模板
// @Description 获取系统邮件模板
// @Tags 管理员-设置
// @Produce json
// @Security Bearer
// @Param name query string true "模板名称"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/settings/email-template [get]
func (h *TemplateHandler) GetEmailTemplate(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		response.BadRequest(c, "模板名称不能为空")
		return
	}

	db := database.Get()
	var template model.MailTemplate
	if err := db.Where("name = ?", name).First(&template).Error; err != nil {
		response.NotFound(c, "模板不存在")
		return
	}

	response.Success(c, template)
}

// GetThemeTemplate 获取主题模板
// @Summary 获取主题模板
// @Description 获取系统主题模板
// @Tags 管理员-设置
// @Produce json
// @Security Bearer
// @Param name query string true "模板名称"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/settings/theme-template [get]
func (h *TemplateHandler) GetThemeTemplate(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		response.BadRequest(c, "模板名称不能为空")
		return
	}

	db := database.Get()
	var theme model.Theme
	if err := db.Where("name = ?", name).First(&theme).Error; err != nil {
		response.NotFound(c, "主题不存在")
		return
	}

	// 解析配置
	var config map[string]interface{}
	if theme.Config != "" {
		// 简单返回原始配置字符串
		config = map[string]interface{}{
			"raw": theme.Config,
		}
	}

	response.Success(c, gin.H{
		"name":   theme.Name,
		"title":  theme.Title,
		"config": config,
	})
}

// TelegramBotHandler Telegram Bot处理器
type TelegramBotHandler struct{}

// NewTelegramBotHandler 创建Telegram Bot处理器
func NewTelegramBotHandler() *TelegramBotHandler {
	return &TelegramBotHandler{}
}

// GetBotInfo 获取Bot信息
// @Summary 获取Bot信息
// @Description 获取Telegram Bot信息
// @Tags 用户-Telegram
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/telegram/bot-info [get]
func (h *TelegramBotHandler) GetBotInfo(c *gin.Context) {
	db := database.Get()

	// 获取Bot配置
	var botTokenSetting, botUsernameSetting model.Setting
	
	botToken := ""
	botUsername := ""
	
	if err := db.Where("`key` = ?", "telegram_bot_token").First(&botTokenSetting).Error; err == nil {
		botToken = botTokenSetting.Value
	}
	
	if err := db.Where("`key` = ?", "telegram_bot_username").First(&botUsernameSetting).Error; err == nil {
		botUsername = botUsernameSetting.Value
	}

	// 如果没有配置，返回空
	if botToken == "" {
		response.Success(c, gin.H{
			"enabled":  false,
			"username": "",
		})
		return
	}

	response.Success(c, gin.H{
		"enabled":  true,
		"username": botUsername,
	})
}

// CheckInviteDetails 检查邀请详情
// @Summary 邀请详情
// @Description 获取用户的邀请详情
// @Tags 用户-邀请
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/invite/details [get]
func (h *TelegramBotHandler) CheckInviteDetails(c *gin.Context) {
	userID, _ := c.Get("user_id")

	db := database.Get()
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		response.BadRequest(c, "用户不存在")
		return
	}

	// 获取邀请码
	var inviteCode model.InviteCode
	if err := db.Where("user_id = ?", userID).First(&inviteCode).Error; err != nil {
		// 没有邀请码，创建一个
		code := generateInviteCode()
		inviteCode = model.InviteCode{
			UserID: user.ID,
			Code:   code,
			Status: 1,
		}
		db.Create(&inviteCode)
	}

	// 获取邀请的人数
	var inviteCount int64
	db.Model(&model.User{}).Where("inviter_id = ?", userID).Count(&inviteCount)

	// 获取佣金统计
	var totalCommission float64
	db.Model(&model.CommissionLog{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalCommission)

	// 获取可提现佣金
	var availableCommission float64
	db.Model(&model.CommissionLog{}).
		Where("user_id = ? AND status = ?", userID, 1).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&availableCommission)

	response.Success(c, gin.H{
		"invite_code":          inviteCode.Code,
		"invite_url":           "/register?code=" + inviteCode.Code,
		"invite_count":         inviteCount,
		"total_commission":     totalCommission,
		"available_commission": availableCommission,
		"commission_balance":   user.Commission,
	})
}

// generateInviteCode 生成邀请码
func generateInviteCode() string {
	// 简单的邀请码生成，实际项目中应该更复杂
	code, _ := utils.GenerateRandomString(8)
	return code
}
