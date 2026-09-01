package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/jwt"
	"xboard-go/pkg/response"
	"xboard-go/pkg/utils"
)

// QuickLoginHandler 快速登录处理器
type QuickLoginHandler struct{}

// NewQuickLoginHandler 创建快速登录处理器
func NewQuickLoginHandler() *QuickLoginHandler {
	return &QuickLoginHandler{}
}

// GetQuickLoginUrl 获取快速登录链接
// @Summary 获取快速登录链接
// @Description 生成快速登录链接，通过邮件发送
// @Tags 认证
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/auth/quick-login-url [post]
func (h *QuickLoginHandler) GetQuickLoginUrl(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	db := database.Get()
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		response.BadRequest(c, "用户不存在")
		return
	}

	// 生成快速登录token
	token, err := utils.GenerateRandomString(64)
	if err != nil {
		response.InternalError(c, "生成token失败")
		return
	}

	// 保存到数据库（有效期15分钟）
	quickToken := model.QuickLoginToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := db.Create(&quickToken).Error; err != nil {
		response.InternalError(c, "保存token失败")
		return
	}

	// 生成登录链接
	loginUrl := "/auth/quick-login?token=" + token

	response.Success(c, gin.H{
		"url":   loginUrl,
		"token": token,
	})
}

// QuickLogin 快速登录
// @Summary 快速登录
// @Description 通过快速登录token登录
// @Tags 认证
// @Produce json
// @Param token query string true "快速登录token"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/quick-login [get]
func (h *QuickLoginHandler) QuickLogin(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.BadRequest(c, "token不能为空")
		return
	}

	db := database.Get()

	// 查找token
	var quickToken model.QuickLoginToken
	if err := db.Where("token = ? AND expires_at > ?", token, time.Now()).
		First(&quickToken).Error; err != nil {
		response.BadRequest(c, "无效或过期的token")
		return
	}

	// 删除已使用的token
	db.Delete(&quickToken)

	// 获取用户
	var user model.User
	if err := db.First(&user, quickToken.UserID).Error; err != nil {
		response.BadRequest(c, "用户不存在")
		return
	}

	// 生成JWT token
	tokenPair, err := jwt.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		response.InternalError(c, "生成token失败")
		return
	}

	// 更新最后登录时间
	db.Model(&user).Updates(map[string]interface{}{
		"last_login_at": time.Now(),
		"last_login_ip": c.ClientIP(),
	})

	response.Success(c, tokenPair)
}

// Token2Login Token直接登录
// @Summary Token直接登录
// @Description 通过token直接登录（管理员功能）
// @Tags 认证
// @Produce json
// @Security Bearer
// @Param user_id query int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/token2login [get]
func (h *QuickLoginHandler) Token2Login(c *gin.Context) {
	// 检查是否是管理员
	role, _ := c.Get("user_role")
	if role != "admin" {
		response.Forbidden(c, "权限不足")
		return
	}

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		response.BadRequest(c, "用户ID不能为空")
		return
	}

	db := database.Get()
	var user model.User
	if err := db.First(&user, userIDStr).Error; err != nil {
		response.BadRequest(c, "用户不存在")
		return
	}

	// 生成JWT token
	tokenPair, err := jwt.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		response.InternalError(c, "生成token失败")
		return
	}

	response.Success(c, tokenPair)
}

// LoginWithMailLink 邮件链接登录
// @Summary 邮件链接登录
// @Description 通过邮件链接登录
// @Tags 认证
// @Produce json
// @Param token query string true "登录token"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/mail-link-login [get]
func (h *QuickLoginHandler) LoginWithMailLink(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.BadRequest(c, "token不能为空")
		return
	}

	db := database.Get()

	// 查找token
	var mailToken model.MailLoginToken
	if err := db.Where("token = ? AND expires_at > ?", token, time.Now()).
		First(&mailToken).Error; err != nil {
		response.BadRequest(c, "无效或过期的token")
		return
	}

	// 删除已使用的token
	db.Delete(&mailToken)

	// 获取用户
	var user model.User
	if err := db.First(&user, mailToken.UserID).Error; err != nil {
		response.BadRequest(c, "用户不存在")
		return
	}

	// 生成JWT token
	tokenPair, err := jwt.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		response.InternalError(c, "生成token失败")
		return
	}

	// 更新最后登录时间
	db.Model(&user).Updates(map[string]interface{}{
		"last_login_at": time.Now(),
		"last_login_ip": c.ClientIP(),
	})

	response.Success(c, tokenPair)
}

// SendMailLoginLink 发送邮件登录链接
// @Summary 发送邮件登录链接
// @Description 发送邮件登录链接给用户
// @Tags 认证
// @Accept json
// @Produce json
// @Param email query string true "邮箱"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/send-mail-login-link [post]
func (h *QuickLoginHandler) SendMailLoginLink(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.BadRequest(c, "邮箱不能为空")
		return
	}

	db := database.Get()
	var user model.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		// 为了安全，不暴露用户是否存在
		response.Success(c, gin.H{"message": "如果邮箱存在，登录链接已发送"})
		return
	}

	// 生成token
	token, err := utils.GenerateRandomString(64)
	if err != nil {
		response.InternalError(c, "生成token失败")
		return
	}

	// 保存token（有效期30分钟）
	mailToken := model.MailLoginToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := db.Create(&mailToken).Error; err != nil {
		response.InternalError(c, "保存token失败")
		return
	}

	// TODO: 发送邮件
	// 这里应该调用邮件服务发送登录链接

	response.Success(c, gin.H{"message": "如果邮箱存在，登录链接已发送"})
}
