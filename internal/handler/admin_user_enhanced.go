package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"
	"xboard-go/pkg/utils"
)

// AdminUserEnhancedHandler 管理员用户增强处理器
type AdminUserEnhancedHandler struct {
}

// NewAdminUserEnhancedHandler 创建管理员用户增强处理器
func NewAdminUserEnhancedHandler() *AdminUserEnhancedHandler {
	return &AdminUserEnhancedHandler{}
}

// SendMailRequest 发送邮件请求
type SendMailRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required,min=1"`
	Subject string `json:"subject" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// SendMailToUsers 给用户发送邮件
// @Summary 发送邮件给用户
// @Description 管理员给指定用户发送邮件
// @Tags 管理员-用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body SendMailRequest true "邮件信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/users/send-mail [post]
func (h *AdminUserEnhancedHandler) SendMailToUsers(c *gin.Context) {
	var req SendMailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	db := database.Get()
	var users []model.User
	if err := db.Where("id IN ?", req.UserIDs).Find(&users).Error; err != nil {
		response.InternalError(c, "查询用户失败")
		return
	}

	// 异步发送邮件
	go func() {
		for _, user := range users {
			// TODO: 调用邮件服务发送邮件
			fmt.Printf("Sending email to %s: %s\n", user.Email, req.Subject)
		}
	}()

	response.Success(c, gin.H{
		"message": fmt.Sprintf("已发送邮件给 %d 位用户", len(users)),
	})
}

// ResetSecretRequest 重置密钥请求
type ResetSecretRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

// ResetUserSecret 重置用户密钥
// @Summary 重置用户密钥
// @Description 重置用户的UUID和订阅Token
// @Tags 管理员-用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body ResetSecretRequest true "用户信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/users/reset-secret [post]
func (h *AdminUserEnhancedHandler) ResetUserSecret(c *gin.Context) {
	var req ResetSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	db := database.Get()
	var user model.User
	if err := db.First(&user, req.UserID).Error; err != nil {
		response.BadRequest(c, "用户不存在")
		return
	}

	// 生成新的UUID和Token
	newUUID, _ := utils.GenerateUUID()
	newToken, _ := utils.GenerateRandomString(32)

	// 更新用户
	if err := db.Model(&user).Updates(map[string]interface{}{
		"uuid":            newUUID,
		"subscribe_token": newToken,
	}).Error; err != nil {
		response.InternalError(c, "重置密钥失败")
		return
	}

	response.Success(c, gin.H{
		"uuid":            newUUID,
		"subscribe_token": newToken,
	})
}

// SetInviteUserRequest 设置邀请人请求
type SetInviteUserRequest struct {
	UserID    uint `json:"user_id" binding:"required"`
	InviterID uint `json:"inviter_id" binding:"required"`
}

// SetInviteUser 设置用户邀请人
// @Summary 设置邀请人
// @Description 管理员设置用户的邀请人
// @Tags 管理员-用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body SetInviteUserRequest true "邀请关系"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/users/set-inviter [post]
func (h *AdminUserEnhancedHandler) SetInviteUser(c *gin.Context) {
	var req SetInviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 不能邀请自己
	if req.UserID == req.InviterID {
		response.BadRequest(c, "不能邀请自己")
		return
	}

	db := database.Get()

	// 检查被邀请人是否存在
	var user model.User
	if err := db.First(&user, req.UserID).Error; err != nil {
		response.BadRequest(c, "用户不存在")
		return
	}

	// 检查邀请人是否存在
	var inviter model.User
	if err := db.First(&inviter, req.InviterID).Error; err != nil {
		response.BadRequest(c, "邀请人不存在")
		return
	}

	// 检查是否已经设置过邀请人
	if user.InviterID != nil && *user.InviterID > 0 {
		response.BadRequest(c, "该用户已有邀请人")
		return
	}

	// 设置邀请人
	if err := db.Model(&user).Update("inviter_id", req.InviterID).Error; err != nil {
		response.InternalError(c, "设置邀请人失败")
		return
	}

	response.Success(c, nil)
}

// TransferBalanceRequest 转移余额请求
type TransferBalanceRequest struct {
	UserID   uint    `json:"user_id" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	TargetID uint    `json:"target_id" binding:"required"`
}

// TransferBalance 转移用户余额
// @Summary 转移余额
// @Description 管理员转移用户余额到另一个用户
// @Tags 管理员-用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body TransferBalanceRequest true "转移信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/users/transfer-balance [post]
func (h *AdminUserEnhancedHandler) TransferBalance(c *gin.Context) {
	var req TransferBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if req.UserID == req.TargetID {
		response.BadRequest(c, "不能转移给自己")
		return
	}

	db := database.Get()

	// 开始事务
	tx := db.Begin()

	// 检查源用户
	var user model.User
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&user, req.UserID).Error; err != nil {
		tx.Rollback()
		response.BadRequest(c, "用户不存在")
		return
	}

	// 检查余额
	if user.Balance < req.Amount {
		tx.Rollback()
		response.BadRequest(c, "余额不足")
		return
	}

	// 检查目标用户
	var target model.User
	if err := tx.First(&target, req.TargetID).Error; err != nil {
		tx.Rollback()
		response.BadRequest(c, "目标用户不存在")
		return
	}

	// 扣除源用户余额
	if err := tx.Model(&user).Update("balance", user.Balance-req.Amount).Error; err != nil {
		tx.Rollback()
		response.InternalError(c, "转移失败")
		return
	}

	// 增加目标用户余额
	if err := tx.Model(&target).Update("balance", target.Balance+req.Amount).Error; err != nil {
		tx.Rollback()
		response.InternalError(c, "转移失败")
		return
	}

	// 提交事务
	tx.Commit()

	response.Success(c, nil)
}
