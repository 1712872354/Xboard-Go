package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// ServerRouteHandler 服务器路由处理器
type ServerRouteHandler struct {
	serverRouteService service.ServerRouteService
}

// NewServerRouteHandler 创建服务器路由处理器
func NewServerRouteHandler(serverRouteService service.ServerRouteService) *ServerRouteHandler {
	return &ServerRouteHandler{
		serverRouteService: serverRouteService,
	}
}

// CreateServerRouteRequest 创建服务器路由请求
type CreateServerRouteRequest struct {
	GroupID uint   `json:"group_id" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Match   string `json:"match" binding:"required"`
	Action  string `json:"action" binding:"required"`
	Target  string `json:"target"`
	Sort    int    `json:"sort"`
}

// CreateServerRoute 创建服务器路由（管理员）
func (h *ServerRouteHandler) CreateServerRoute(c *gin.Context) {
	var req CreateServerRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	route, err := h.serverRouteService.Create(req.GroupID, req.Name, req.Match, req.Action, req.Target, req.Sort)
	if err != nil {
		response.InternalError(c, "创建服务器路由失败："+err.Error())
		return
	}

	response.Success(c, route)
}

// GetServerRoute 获取服务器路由详情（管理员）
func (h *ServerRouteHandler) GetServerRoute(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的路由ID")
		return
	}

	route, err := h.serverRouteService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "服务器路由不存在")
		return
	}

	response.Success(c, route)
}

// UpdateServerRouteRequest 更新服务器路由请求
type UpdateServerRouteRequest struct {
	GroupID uint   `json:"group_id" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Match   string `json:"match" binding:"required"`
	Action  string `json:"action" binding:"required"`
	Target  string `json:"target"`
	Sort    int    `json:"sort"`
	Status  int    `json:"status"`
}

// UpdateServerRoute 更新服务器路由（管理员）
func (h *ServerRouteHandler) UpdateServerRoute(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的路由ID")
		return
	}

	var req UpdateServerRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	route, err := h.serverRouteService.Update(uint(id), req.GroupID, req.Name, req.Match, req.Action, req.Target, req.Sort, req.Status)
	if err != nil {
		response.InternalError(c, "更新服务器路由失败："+err.Error())
		return
	}

	response.Success(c, route)
}

// DeleteServerRoute 删除服务器路由（管理员）
func (h *ServerRouteHandler) DeleteServerRoute(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的路由ID")
		return
	}

	if err := h.serverRouteService.Delete(uint(id)); err != nil {
		response.InternalError(c, "删除服务器路由失败")
		return
	}

	response.Success(c, nil)
}

// ListServerRoutes 服务器路由列表（管理员）
func (h *ServerRouteHandler) ListServerRoutes(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	groupIDStr := c.DefaultQuery("group_id", "0")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		groupID = 0
	}

	routes, total, err := h.serverRouteService.List(page, pageSize, uint(groupID))
	if err != nil {
		response.InternalError(c, "获取服务器路由列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  routes,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// ListServerRoutesByGroup 根据分组获取路由列表
func (h *ServerRouteHandler) ListServerRoutesByGroup(c *gin.Context) {
	groupIDStr := c.Param("group_id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的分组ID")
		return
	}

	routes, err := h.serverRouteService.ListByGroup(uint(groupID))
	if err != nil {
		response.InternalError(c, "获取服务器路由列表失败")
		return
	}

	response.Success(c, routes)
}
