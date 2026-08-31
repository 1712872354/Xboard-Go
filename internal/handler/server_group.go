package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// ServerGroupHandler 服务器分组处理器
type ServerGroupHandler struct {
	serverGroupService service.ServerGroupService
}

// NewServerGroupHandler 创建服务器分组处理器
func NewServerGroupHandler(serverGroupService service.ServerGroupService) *ServerGroupHandler {
	return &ServerGroupHandler{
		serverGroupService: serverGroupService,
	}
}

// CreateServerGroupRequest 创建服务器分组请求
type CreateServerGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	PlanIDs     string `json:"plan_ids"`
	Sort        int    `json:"sort"`
}

// CreateServerGroup 创建服务器分组（管理员）
func (h *ServerGroupHandler) CreateServerGroup(c *gin.Context) {
	var req CreateServerGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	group, err := h.serverGroupService.Create(req.Name, req.Description, req.PlanIDs, req.Sort)
	if err != nil {
		response.InternalError(c, "创建服务器分组失败："+err.Error())
		return
	}

	response.Success(c, group)
}

// GetServerGroup 获取服务器分组详情（管理员）
func (h *ServerGroupHandler) GetServerGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的分组ID")
		return
	}

	group, err := h.serverGroupService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "服务器分组不存在")
		return
	}

	response.Success(c, group)
}

// UpdateServerGroupRequest 更新服务器分组请求
type UpdateServerGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	PlanIDs     string `json:"plan_ids"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status"`
}

// UpdateServerGroup 更新服务器分组（管理员）
func (h *ServerGroupHandler) UpdateServerGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的分组ID")
		return
	}

	var req UpdateServerGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	group, err := h.serverGroupService.Update(uint(id), req.Name, req.Description, req.PlanIDs, req.Sort, req.Status)
	if err != nil {
		response.InternalError(c, "更新服务器分组失败："+err.Error())
		return
	}

	response.Success(c, group)
}

// DeleteServerGroup 删除服务器分组（管理员）
func (h *ServerGroupHandler) DeleteServerGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的分组ID")
		return
	}

	if err := h.serverGroupService.Delete(uint(id)); err != nil {
		response.InternalError(c, "删除服务器分组失败")
		return
	}

	response.Success(c, nil)
}

// ListServerGroups 服务器分组列表（管理员）
func (h *ServerGroupHandler) ListServerGroups(c *gin.Context) {
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

	groups, total, err := h.serverGroupService.List(page, pageSize)
	if err != nil {
		response.InternalError(c, "获取服务器分组列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  groups,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// ListAllServerGroups 获取所有服务器分组
func (h *ServerGroupHandler) ListAllServerGroups(c *gin.Context) {
	groups, err := h.serverGroupService.ListAll()
	if err != nil {
		response.InternalError(c, "获取服务器分组列表失败")
		return
	}

	response.Success(c, groups)
}
