package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/middleware"
	"xboard-go/internal/model"
	"xboard-go/internal/service"
	"xboard-go/pkg/database"
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

// ResetMailTemplate 恢复默认模板（管理员）
func (h *MailTemplateHandler) ResetMailTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	template, err := h.mailTemplateService.ResetMailTemplate(uint(id))
	if err != nil {
		response.InternalError(c, "恢复默认模板失败："+err.Error())
		return
	}

	response.Success(c, template)
}

// TestMailTemplateRequest 发送测试邮件请求
type TestMailTemplateRequest struct {
	Email string `json:"email"`
}

// TestMailTemplate 发送测试邮件（管理员）
func (h *MailTemplateHandler) TestMailTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	var req TestMailTemplateRequest
	_ = c.ShouldBindJSON(&req)

	email := req.Email
	if email == "" {
		// 使用当前管理员邮箱
		userID := middleware.GetUserID(c)
		var user model.User
		if err := database.Get().First(&user, userID).Error; err != nil {
			response.BadRequest(c, "无法获取管理员邮箱")
			return
		}
		email = user.Email
	}

	if err := h.mailTemplateService.TestMailTemplate(uint(id), email); err != nil {
		response.InternalError(c, "发送测试邮件失败："+err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "测试邮件已发送至 " + email,
	})
}
