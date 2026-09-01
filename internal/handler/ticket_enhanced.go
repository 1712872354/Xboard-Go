package handler

import (
	"github.com/gin-gonic/gin"
	"xboard-go/internal/middleware"
	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"

	"gorm.io/gorm"
)

// TicketEnhancedHandler 工单增强处理器
type TicketEnhancedHandler struct {
	db *gorm.DB
}

// NewTicketEnhancedHandler 创建工单增强处理器
func NewTicketEnhancedHandler() *TicketEnhancedHandler {
	return &TicketEnhancedHandler{
		db: database.Get(),
	}
}

// WithdrawTicket 撤回工单
// @Summary 撤回工单
// @Description 用户撤回待处理的工单
// @Tags 工单
// @Produce json
// @Security Bearer
// @Param id path int true "工单ID"
// @Success 200 {object} response.Response
// @Router /api/v1/tickets/{id}/withdraw [post]
func (h *TicketEnhancedHandler) WithdrawTicket(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		response.BadRequest(c, "无效的工单ID")
		return
	}

	// 查询工单
	var ticket model.Ticket
	if err := h.db.First(&ticket, id).Error; err != nil {
		response.BadRequest(c, "工单不存在")
		return
	}

	// 验证归属
	if ticket.UserID != userID {
		response.BadRequest(c, "无权操作此工单")
		return
	}

	// 只有待处理状态才能撤回
	if ticket.Status != 0 {
		response.BadRequest(c, "只能撤回待处理的工单")
		return
	}

	// 删除工单及其回复
	tx := h.db.Begin()
	if err := tx.Where("ticket_id = ?", id).Delete(&model.TicketMessage{}).Error; err != nil {
		tx.Rollback()
		response.InternalError(c, "撤回失败")
		return
	}
	if err := tx.Delete(&ticket).Error; err != nil {
		tx.Rollback()
		response.InternalError(c, "撤回失败")
		return
	}
	tx.Commit()

	response.Success(c, nil)
}

// GetTicketMessages 获取工单消息列表
// @Summary 获取工单消息
// @Description 获取工单的所有消息（包括初始内容和回复）
// @Tags 工单
// @Produce json
// @Security Bearer
// @Param id path int true "工单ID"
// @Success 200 {object} response.Response
// @Router /api/v1/tickets/{id}/messages [get]
func (h *TicketEnhancedHandler) GetTicketMessages(c *gin.Context) {
	userID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)

	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		response.BadRequest(c, "无效的工单ID")
		return
	}

	// 查询工单
	var ticket model.Ticket
	if err := h.db.First(&ticket, id).Error; err != nil {
		response.BadRequest(c, "工单不存在")
		return
	}

	// 验证权限
	if !isAdmin && ticket.UserID != userID {
		response.BadRequest(c, "无权查看此工单")
		return
	}

	// 获取消息列表
	var messages []model.TicketMessage
	if err := h.db.Where("ticket_id = ?", id).
		Preload("User").
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		response.InternalError(c, "获取消息失败")
		return
	}

	// 构建响应
	var result []gin.H
	for _, msg := range messages {
		result = append(result, gin.H{
			"id":         msg.ID,
			"content":    msg.Content,
			"is_admin":   msg.IsAdmin,
			"user_id":    msg.UserID,
			"user_email": msg.User.Email,
			"created_at": msg.CreatedAt,
		})
	}

	if result == nil {
		result = []gin.H{}
	}

	response.Success(c, result)
}
