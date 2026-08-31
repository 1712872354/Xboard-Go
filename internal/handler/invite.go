package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// InviteHandler 邀请码处理器
type InviteHandler struct {
	inviteService service.InviteService
}

// NewInviteHandler 创建邀请码处理器
func NewInviteHandler(inviteService service.InviteService) *InviteHandler {
	return &InviteHandler{
		inviteService: inviteService,
	}
}

// CreateInviteCodeRequest 创建邀请码请求
type CreateInviteCodeRequest struct {
	UserID     uint    `json:"user_id" binding:"required"`
	Commission float64 `json:"commission"`
	LimitCount int     `json:"limit_count"`
}

// CreateInviteCode 创建邀请码（管理员）
func (h *InviteHandler) CreateInviteCode(c *gin.Context) {
	var req CreateInviteCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	code, err := h.inviteService.Create(req.UserID, req.Commission, req.LimitCount)
	if err != nil {
		response.InternalError(c, "创建邀请码失败："+err.Error())
		return
	}

	response.Success(c, code)
}

// GetInviteCode 获取邀请码详情（管理员）
func (h *InviteHandler) GetInviteCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的邀请码ID")
		return
	}

	code, err := h.inviteService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "邀请码不存在")
		return
	}

	response.Success(c, code)
}

// UpdateInviteCodeRequest 更新邀请码请求
type UpdateInviteCodeRequest struct {
	Commission float64 `json:"commission"`
	LimitCount int     `json:"limit_count"`
	Status     int     `json:"status"`
}

// UpdateInviteCode 更新邀请码（管理员）
func (h *InviteHandler) UpdateInviteCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的邀请码ID")
		return
	}

	var req UpdateInviteCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	code, err := h.inviteService.Update(uint(id), req.Commission, req.LimitCount, req.Status)
	if err != nil {
		response.InternalError(c, "更新邀请码失败："+err.Error())
		return
	}

	response.Success(c, code)
}

// DeleteInviteCode 删除邀请码（管理员）
func (h *InviteHandler) DeleteInviteCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的邀请码ID")
		return
	}

	if err := h.inviteService.Delete(uint(id)); err != nil {
		response.InternalError(c, "删除邀请码失败")
		return
	}

	response.Success(c, nil)
}

// ListInviteCodes 邀请码列表（管理员）
func (h *InviteHandler) ListInviteCodes(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	codes, total, err := h.inviteService.List(page, pageSize)
	if err != nil {
		response.InternalError(c, "获取邀请码列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  codes,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// GetUserInviteCode 获取用户自己的邀请码
func (h *InviteHandler) GetUserInviteCode(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	code, err := h.inviteService.GetByUserID(userID.(uint))
	if err != nil {
		response.NotFound(c, "邀请码不存在")
		return
	}

	response.Success(c, code)
}

// UseInviteCodeRequest 使用邀请码请求
type UseInviteCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

// UseInviteCode 使用邀请码
func (h *InviteHandler) UseInviteCode(c *gin.Context) {
	var req UseInviteCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	if err := h.inviteService.UseCode(req.Code, userID.(uint)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetCommissionStats 获取佣金统计
func (h *InviteHandler) GetCommissionStats(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	total, err := h.inviteService.GetTotalCommission(userID.(uint))
	if err != nil {
		response.InternalError(c, "获取佣金统计失败")
		return
	}

	pending, err := h.inviteService.GetPendingCommission(userID.(uint))
	if err != nil {
		response.InternalError(c, "获取佣金统计失败")
		return
	}

	response.Success(c, gin.H{
		"total":   total,
		"pending": pending,
	})
}

// ListCommissionLogs 佣金记录列表
func (h *InviteHandler) ListCommissionLogs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	logs, total, err := h.inviteService.GetCommissionLogs(page, pageSize, userID.(uint))
	if err != nil {
		response.InternalError(c, "获取佣金记录失败")
		return
	}

	response.Success(c, gin.H{
		"list":  logs,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// SettleCommission 结算佣金（管理员）
func (h *InviteHandler) SettleCommission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的佣金记录ID")
		return
	}

	if err := h.inviteService.SettleCommission(uint(id)); err != nil {
		response.InternalError(c, "结算佣金失败："+err.Error())
		return
	}

	response.Success(c, nil)
}

// ListCommissionLogsAdmin 佣金记录列表（管理员）
func (h *InviteHandler) ListCommissionLogsAdmin(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	userIDStr := c.DefaultQuery("user_id", "0")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		userID = 0
	}

	logs, total, err := h.inviteService.GetCommissionLogs(page, pageSize, uint(userID))
	if err != nil {
		response.InternalError(c, "获取佣金记录失败")
		return
	}

	response.Success(c, gin.H{
		"list":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// WithdrawCommission 佣金提现
func (h *InviteHandler) WithdrawCommission(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	amount, err := h.inviteService.WithdrawCommission(userID.(uint))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"amount":  amount,
		"message": "佣金已转入余额",
	})
}
