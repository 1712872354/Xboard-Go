package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// NodeTemplateHandler 节点模板处理器
type NodeTemplateHandler struct {
	templateService service.NodeTemplateService
}

// NewNodeTemplateHandler 创建节点模板处理器
func NewNodeTemplateHandler(templateService service.NodeTemplateService) *NodeTemplateHandler {
	return &NodeTemplateHandler{
		templateService: templateService,
	}
}

// CreateNodeTemplateRequest 创建节点模板请求
type CreateNodeTemplateRequest struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	ServerInfo  string `json:"server_info"`
	Description string `json:"description"`
}

// CreateNodeTemplate 创建节点模板
func (h *NodeTemplateHandler) CreateNodeTemplate(c *gin.Context) {
	var req CreateNodeTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	template, err := h.templateService.CreateNodeTemplate(req.Name, req.Type, req.ServerInfo, req.Description)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, template)
}

// GetNodeTemplate 获取节点模板详情
func (h *NodeTemplateHandler) GetNodeTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid template ID")
		return
	}

	template, err := h.templateService.GetNodeTemplateByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, template)
}

// UpdateNodeTemplateRequest 更新节点模板请求
type UpdateNodeTemplateRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	ServerInfo  string `json:"server_info"`
	Description string `json:"description"`
}

// UpdateNodeTemplate 更新节点模板
func (h *NodeTemplateHandler) UpdateNodeTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid template ID")
		return
	}

	var req UpdateNodeTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	template, err := h.templateService.UpdateNodeTemplate(uint(id), req.Name, req.Type, req.ServerInfo, req.Description)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, template)
}

// DeleteNodeTemplate 删除节点模板
func (h *NodeTemplateHandler) DeleteNodeTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid template ID")
		return
	}

	if err := h.templateService.DeleteNodeTemplate(uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ListNodeTemplates 节点模板列表
func (h *NodeTemplateHandler) ListNodeTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	templates, total, err := h.templateService.ListNodeTemplates(page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      templates,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
