package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// GiftCardHandler 礼品卡处理器
type GiftCardHandler struct {
	giftCardService service.GiftCardService
}

// NewGiftCardHandler 创建礼品卡处理器
func NewGiftCardHandler(giftCardService service.GiftCardService) *GiftCardHandler {
	return &GiftCardHandler{
		giftCardService: giftCardService,
	}
}

// CreateTemplateRequest 创建礼品卡模板请求
type CreateTemplateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Type        int     `json:"type" binding:"required"`
	Value       float64 `json:"value"`
	Traffic     int64   `json:"traffic"`
	Duration    int     `json:"duration"`
	PlanID      *uint   `json:"plan_id"`
	Price       float64 `json:"price"`
}

// CreateTemplate 创建礼品卡模板（管理员）
func (h *GiftCardHandler) CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	template, err := h.giftCardService.CreateTemplate(
		req.Name, req.Description, req.Type, req.Value,
		req.Traffic, req.Duration, req.PlanID, req.Price,
	)
	if err != nil {
		response.InternalError(c, "创建礼品卡模板失败："+err.Error())
		return
	}

	response.Success(c, template)
}

// GetTemplate 获取礼品卡模板详情（管理员）
func (h *GiftCardHandler) GetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	template, err := h.giftCardService.GetTemplateByID(uint(id))
	if err != nil {
		response.NotFound(c, "模板不存在")
		return
	}

	response.Success(c, template)
}

// UpdateTemplateRequest 更新礼品卡模板请求
type UpdateTemplateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Type        int     `json:"type" binding:"required"`
	Value       float64 `json:"value"`
	Traffic     int64   `json:"traffic"`
	Duration    int     `json:"duration"`
	PlanID      *uint   `json:"plan_id"`
	Price       float64 `json:"price"`
	Status      int     `json:"status"`
}

// UpdateTemplate 更新礼品卡模板（管理员）
func (h *GiftCardHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	template, err := h.giftCardService.UpdateTemplate(
		uint(id), req.Name, req.Description, req.Type, req.Value,
		req.Traffic, req.Duration, req.PlanID, req.Price, req.Status,
	)
	if err != nil {
		response.InternalError(c, "更新礼品卡模板失败："+err.Error())
		return
	}

	response.Success(c, template)
}

// DeleteTemplate 删除礼品卡模板（管理员）
func (h *GiftCardHandler) DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	if err := h.giftCardService.DeleteTemplate(uint(id)); err != nil {
		response.InternalError(c, "删除礼品卡模板失败")
		return
	}

	response.Success(c, nil)
}

// ListTemplates 礼品卡模板列表（管理员）
func (h *GiftCardHandler) ListTemplates(c *gin.Context) {
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

	templates, total, err := h.giftCardService.ListTemplates(page, pageSize)
	if err != nil {
		response.InternalError(c, "获取礼品卡模板列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  templates,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// GenerateCodesRequest 生成礼品码请求
type GenerateCodesRequest struct {
	TemplateID uint `json:"template_id" binding:"required"`
	Count      int  `json:"count" binding:"required"`
}

// GenerateCodes 批量生成礼品码（管理员）
func (h *GiftCardHandler) GenerateCodes(c *gin.Context) {
	var req GenerateCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	codes, err := h.giftCardService.GenerateCodes(req.TemplateID, req.Count)
	if err != nil {
		response.InternalError(c, "生成礼品码失败："+err.Error())
		return
	}

	response.Success(c, gin.H{
		"codes": codes,
		"count": len(codes),
	})
}

// GetCode 获取礼品码详情（管理员）
func (h *GiftCardHandler) GetCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的礼品码ID")
		return
	}

	code, err := h.giftCardService.GetCodeByID(uint(id))
	if err != nil {
		response.NotFound(c, "礼品码不存在")
		return
	}

	response.Success(c, code)
}

// DeleteCode 删除礼品码（管理员）
func (h *GiftCardHandler) DeleteCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的礼品码ID")
		return
	}

	if err := h.giftCardService.DeleteCode(uint(id)); err != nil {
		response.InternalError(c, "删除礼品码失败")
		return
	}

	response.Success(c, nil)
}

// ListCodes 礼品码列表（管理员）
func (h *GiftCardHandler) ListCodes(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	templateIDStr := c.DefaultQuery("template_id", "0")
	statusStr := c.DefaultQuery("status", "-1")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		templateID = 0
	}

	status, err := strconv.Atoi(statusStr)
	if err != nil {
		status = -1
	}

	codes, total, err := h.giftCardService.ListCodes(page, pageSize, uint(templateID), status)
	if err != nil {
		response.InternalError(c, "获取礼品码列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  codes,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// UseCodeRequest 使用礼品码请求
type UseCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

// UseCode 使用礼品码（用户端）
func (h *GiftCardHandler) UseCode(c *gin.Context) {
	var req UseCodeRequest
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

	usage, err := h.giftCardService.UseCode(req.Code, userID.(uint))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, usage)
}
