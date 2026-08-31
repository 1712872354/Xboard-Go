package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// MailTemplateHandler 邮件模板处理器
type MailTemplateHandler struct {
	mailTemplateService service.MailTemplateService
}

// NewMailTemplateHandler 创建邮件模板处理器
func NewMailTemplateHandler(mailTemplateService service.MailTemplateService) *MailTemplateHandler {
	return &MailTemplateHandler{
		mailTemplateService: mailTemplateService,
	}
}

// CreateMailTemplateRequest 创建邮件模板请求
type CreateMailTemplateRequest struct {
	Name     string `json:"name" binding:"required"`
	Subject  string `json:"subject" binding:"required"`
	Body     string `json:"body" binding:"required"`
	Language string `json:"language"`
	Remark   string `json:"remark"`
}

// CreateMailTemplate 创建邮件模板（管理员）
func (h *MailTemplateHandler) CreateMailTemplate(c *gin.Context) {
	var req CreateMailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if req.Language == "" {
		req.Language = "zh-CN"
	}

	template, err := h.mailTemplateService.Create(req.Name, req.Subject, req.Body, req.Language, req.Remark)
	if err != nil {
		response.InternalError(c, "创建邮件模板失败："+err.Error())
		return
	}

	response.Success(c, template)
}

// GetMailTemplate 获取邮件模板详情（管理员）
func (h *MailTemplateHandler) GetMailTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	template, err := h.mailTemplateService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "模板不存在")
		return
	}

	response.Success(c, template)
}

// UpdateMailTemplateRequest 更新邮件模板请求
type UpdateMailTemplateRequest struct {
	Name     string `json:"name" binding:"required"`
	Subject  string `json:"subject" binding:"required"`
	Body     string `json:"body" binding:"required"`
	Language string `json:"language"`
	Remark   string `json:"remark"`
}

// UpdateMailTemplate 更新邮件模板（管理员）
func (h *MailTemplateHandler) UpdateMailTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	var req UpdateMailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if req.Language == "" {
		req.Language = "zh-CN"
	}

	template, err := h.mailTemplateService.Update(uint(id), req.Name, req.Subject, req.Body, req.Language, req.Remark)
	if err != nil {
		response.InternalError(c, "更新邮件模板失败："+err.Error())
		return
	}

	response.Success(c, template)
}

// DeleteMailTemplate 删除邮件模板（管理员）
func (h *MailTemplateHandler) DeleteMailTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	if err := h.mailTemplateService.Delete(uint(id)); err != nil {
		response.InternalError(c, "删除邮件模板失败")
		return
	}

	response.Success(c, nil)
}

// ListMailTemplates 邮件模板列表（管理员）
func (h *MailTemplateHandler) ListMailTemplates(c *gin.Context) {
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

	templates, total, err := h.mailTemplateService.List(page, pageSize)
	if err != nil {
		response.InternalError(c, "获取邮件模板列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  templates,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}
