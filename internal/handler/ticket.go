package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/middleware"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// TicketHandler 工单处理器
type TicketHandler struct {
	ticketService service.TicketService
}

// NewTicketHandler 创建工单处理器
func NewTicketHandler(ticketService service.TicketService) *TicketHandler {
	return &TicketHandler{
		ticketService: ticketService,
	}
}

// CreateTicketRequest 创建工单请求
type CreateTicketRequest struct {
	Subject  string `json:"subject" binding:"required,max=200"`
	Content  string `json:"content" binding:"required"`
	Category int    `json:"category"`
	Priority int    `json:"priority"`
}

// CreateTicket 创建工单
// @Summary 创建工单
// @Description 用户提交新工单
// @Tags 工单
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateTicketRequest true "工单信息"
// @Success 200 {object} response.Response{data=model.Ticket}
// @Router /api/v1/tickets [post]
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	ticket, err := h.ticketService.CreateTicket(userID, req.Subject, req.Content, req.Category, req.Priority)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, ticket)
}

// GetTicket 获取工单详情
// @Summary 工单详情
// @Description 获取工单详情及回复列表
// @Tags 工单
// @Produce json
// @Security Bearer
// @Param id path int true "工单ID"
// @Success 200 {object} response.Response
// @Router /api/v1/tickets/{id} [get]
func (h *TicketHandler) GetTicket(c *gin.Context) {
	userID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)

	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	ticket, replies, err := h.ticketService.GetTicket(id, userID, isAdmin)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"ticket":  ticket,
		"replies": replies,
	})
}

// ReplyRequest 回复请求
type ReplyRequest struct {
	Content string `json:"content" binding:"required"`
}

// Reply 回复工单
// @Summary 回复工单
// @Description 用户或管理员回复工单
// @Tags 工单
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "工单ID"
// @Param request body ReplyRequest true "回复内容"
// @Success 200 {object} response.Response
// @Router /api/v1/tickets/{id}/reply [post]
func (h *TicketHandler) Reply(c *gin.Context) {
	userID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)

	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	var req ReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if err := h.ticketService.Reply(id, userID, req.Content, isAdmin); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// CloseTicket 关闭工单
// @Summary 关闭工单
// @Description 用户或管理员关闭工单
// @Tags 工单
// @Produce json
// @Security Bearer
// @Param id path int true "工单ID"
// @Success 200 {object} response.Response
// @Router /api/v1/tickets/{id}/close [post]
func (h *TicketHandler) CloseTicket(c *gin.Context) {
	userID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)

	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	if err := h.ticketService.CloseTicket(id, userID, isAdmin); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ListUserTickets 用户工单列表
// @Summary 我的工单
// @Description 获取当前用户的工单列表
// @Tags 工单
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query int false "状态筛选 (-1=全部, 0=待处理, 1=已回复, 2=已关闭)" default(-1)
// @Success 200 {object} response.Response
// @Router /api/v1/tickets [get]
func (h *TicketHandler) ListUserTickets(c *gin.Context) {
	userID := middleware.GetUserID(c)

	page := atoi(c.DefaultQuery("page", "1"))
	pageSize := atoi(c.DefaultQuery("page_size", "20"))
	status := atoi(c.DefaultQuery("status", "-1"))

	tickets, total, err := h.ticketService.ListUserTickets(userID, page, pageSize, status)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      tickets,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetUserStats 用户工单统计
// @Summary 我的工单统计
// @Description 获取当前用户的工单统计数据
// @Tags 工单
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=service.TicketStats}
// @Router /api/v1/tickets/stats [get]
func (h *TicketHandler) GetUserStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)

	stats, err := h.ticketService.GetStats(userID, isAdmin)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

// ===== 管理员工单接口 =====

// ListAllTickets 所有工单列表
// @Summary 工单列表（管理员）
// @Description 获取所有工单列表（管理员权限）
// @Tags 管理员-工单
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query int false "状态筛选 (-1=全部)" default(-1)
// @Param category query int false "分类筛选 (-1=全部)" default(-1)
// @Success 200 {object} response.Response
// @Router /api/v1/admin/tickets [get]
func (h *TicketHandler) ListAllTickets(c *gin.Context) {
	page := atoi(c.DefaultQuery("page", "1"))
	pageSize := atoi(c.DefaultQuery("page_size", "20"))
	status := atoi(c.DefaultQuery("status", "-1"))
	category := atoi(c.DefaultQuery("category", "-1"))

	tickets, total, err := h.ticketService.ListAllTickets(page, pageSize, status, category)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      tickets,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteTicket 删除工单
// @Summary 删除工单（管理员）
// @Description 删除工单（管理员权限）
// @Tags 管理员-工单
// @Produce json
// @Security Bearer
// @Param id path int true "工单ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/tickets/{id} [delete]
func (h *TicketHandler) DeleteTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	if err := h.ticketService.DeleteTicket(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// 辅助函数
func parseUint(s string) (uint, error) {
	var n uint
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func atoi(s string) int {
	n := 0
	fmt.Sscanf(s, "%d", &n)
	return n
}
