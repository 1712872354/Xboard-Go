package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// AdminAuditLogHandler 管理员审计日志处理器
type AdminAuditLogHandler struct {
	auditLogService service.AdminAuditLogService
}

// NewAdminAuditLogHandler 创建管理员审计日志处理器
func NewAdminAuditLogHandler(auditLogService service.AdminAuditLogService) *AdminAuditLogHandler {
	return &AdminAuditLogHandler{
		auditLogService: auditLogService,
	}
}

// GetAuditLog 获取审计日志详情（管理员）
func (h *AdminAuditLogHandler) GetAuditLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的日志ID")
		return
	}

	log, err := h.auditLogService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "审计日志不存在")
		return
	}

	response.Success(c, log)
}

// ListAuditLogs 审计日志列表（管理员）
func (h *AdminAuditLogHandler) ListAuditLogs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	userIDStr := c.DefaultQuery("user_id", "0")
	action := c.Query("action")

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

	logs, total, err := h.auditLogService.List(page, pageSize, uint(userID), action)
	if err != nil {
		response.InternalError(c, "获取审计日志列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  logs,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// DeleteAuditLog 删除审计日志（管理员）
func (h *AdminAuditLogHandler) DeleteAuditLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的日志ID")
		return
	}

	if err := h.auditLogService.Delete(uint(id)); err != nil {
		response.InternalError(c, "删除审计日志失败")
		return
	}

	response.Success(c, nil)
}
