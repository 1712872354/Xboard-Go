package handler

import (
	"strconv"

	"xboard-go/internal/service"
	"xboard-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// NotifyNodeConfigChange is called after a node is updated via the admin API.
// It is set by cmd/server/main.go to bridge the HTTP handler to the gRPC
// broadcaster, avoiding a circular import between handler and grpc packages.
// Parameters: nodeID (uint), name, typ, address (string), port (int), serverInfo (string), rate (float64), groupID, parentID (uint)
var NotifyNodeConfigChange func(nodeID uint, name, typ, address string, port int, serverInfo string, rate float64, groupID, parentID uint)

// NodeHandler 节点处理器
type NodeHandler struct {
	nodeService service.NodeService
}

// NewNodeHandler 创建节点处理器
func NewNodeHandler(nodeService service.NodeService) *NodeHandler {
	return &NodeHandler{
		nodeService: nodeService,
	}
}

// CreateNodeRequest 创建节点请求
type CreateNodeRequest struct {
	Name                string  `json:"name" binding:"required"`
	Type                string  `json:"type" binding:"required"`
	Address             string  `json:"address" binding:"required"`
	Port                int     `json:"port" binding:"required,min=1,max=65535"`
	ServerInfo          string  `json:"server_info"`
	GroupID             uint    `json:"group_id"`
	Rate                float64 `json:"rate"`
	ParentID            uint    `json:"parent_id"`
	Sort                int     `json:"sort"`
	Show                *int    `json:"show"` // 使用指针区分0值和未设置
	Tags                string  `json:"tags"`
	HealthCheckPort     int     `json:"health_check_port"`
	HealthCheckInterval int     `json:"health_check_interval"`
	HealthCheckTimeout  int     `json:"health_check_timeout"`
	HealthCheckType     string  `json:"health_check_type"`
}

// CreateNode 创建节点
// @Summary 创建节点
// @Description 创建新的节点（管理员权限）
// @Tags 管理员-节点
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateNodeRequest true "节点信息"
// @Success 200 {object} response.Response{data=model.Node}
// @Router /api/v1/admin/nodes [post]
func (h *NodeHandler) CreateNode(c *gin.Context) {
	var req CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	showVal := 1
	if req.Show != nil {
		showVal = *req.Show
	}

	node, err := h.nodeService.CreateNodeEx(service.CreateNodeRequest{
		Name:                req.Name,
		Type:                req.Type,
		Address:             req.Address,
		Port:                req.Port,
		ServerInfo:          req.ServerInfo,
		GroupID:             req.GroupID,
		Rate:                req.Rate,
		ParentID:            req.ParentID,
		Sort:                req.Sort,
		Show:                showVal,
		Tags:                req.Tags,
		HealthCheckPort:     req.HealthCheckPort,
		HealthCheckInterval: req.HealthCheckInterval,
		HealthCheckTimeout:  req.HealthCheckTimeout,
		HealthCheckType:     req.HealthCheckType,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, node)
}

// GetNode 获取节点详情
// @Summary 获取节点详情
// @Description 获取指定节点的详细信息
// @Tags 管理员-节点
// @Produce json
// @Security Bearer
// @Param id path int true "节点ID"
// @Success 200 {object} response.Response{data=model.Node}
// @Router /api/v1/admin/nodes/{id} [get]
func (h *NodeHandler) GetNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid node ID")
		return
	}

	node, err := h.nodeService.GetNodeByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, node)
}

// UpdateNodeRequest 更新节点请求
type UpdateNodeRequest struct {
	Name                string  `json:"name"`
	Type                string  `json:"type"`
	Address             string  `json:"address"`
	Port                int     `json:"port"`
	ServerInfo          string  `json:"server_info"`
	GroupID             uint    `json:"group_id"`
	Rate                float64 `json:"rate"`
	Status              int     `json:"status"`
	ParentID            uint    `json:"parent_id"`
	Sort                *int    `json:"sort"`
	Show                *int    `json:"show"`
	Tags                string  `json:"tags"`
	HealthCheckPort     int     `json:"health_check_port"`
	HealthCheckInterval int     `json:"health_check_interval"`
	HealthCheckTimeout  int     `json:"health_check_timeout"`
	HealthCheckType     string  `json:"health_check_type"`
}

// UpdateNode 更新节点
// @Summary 更新节点
// @Description 更新节点信息（管理员权限）
// @Tags 管理员-节点
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "节点ID"
// @Param request body UpdateNodeRequest true "更新信息"
// @Success 200 {object} response.Response{data=model.Node}
// @Router /api/v1/admin/nodes/{id} [put]
func (h *NodeHandler) UpdateNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid node ID")
		return
	}

	var req UpdateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	node, err := h.nodeService.UpdateNodeEx(uint(id), service.UpdateNodeRequest{
		Name:                req.Name,
		Type:                req.Type,
		Address:             req.Address,
		Port:                req.Port,
		ServerInfo:          req.ServerInfo,
		GroupID:             req.GroupID,
		Rate:                req.Rate,
		Status:              req.Status,
		ParentID:            req.ParentID,
		Sort:                req.Sort,
		Show:                req.Show,
		Tags:                req.Tags,
		HealthCheckPort:     req.HealthCheckPort,
		HealthCheckInterval: req.HealthCheckInterval,
		HealthCheckTimeout:  req.HealthCheckTimeout,
		HealthCheckType:     req.HealthCheckType,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Push config update to connected gRPC nodes.
	if NotifyNodeConfigChange != nil {
		go NotifyNodeConfigChange(uint(id), node.Name, node.Type, node.Address, node.Port, node.ServerInfo, node.Rate, node.GroupID, node.ParentID)
	}

	response.Success(c, node)
}

// DeleteNode 删除节点
// @Summary 删除节点
// @Description 删除节点（管理员权限）
// @Tags 管理员-节点
// @Produce json
// @Security Bearer
// @Param id path int true "节点ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/nodes/{id} [delete]
func (h *NodeHandler) DeleteNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid node ID")
		return
	}

	if err := h.nodeService.DeleteNode(uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ListNodes 节点列表
// @Summary 节点列表
// @Description 获取节点列表（管理员权限）
// @Tags 管理员-节点
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param group_id query int false "节点组ID" default(0)
// @Success 200 {object} response.Response
// @Router /api/v1/admin/nodes [get]
func (h *NodeHandler) ListNodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	groupID, _ := strconv.ParseUint(c.DefaultQuery("group_id", "0"), 10, 32)

	nodes, total, err := h.nodeService.ListNodes(page, pageSize, uint(groupID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      nodes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateNodeStatusRequest 更新节点状态请求
type UpdateNodeStatusRequest struct {
	Status int `json:"status" binding:"required,oneof=0 1 2"`
}

// UpdateNodeStatus 更新节点状态
// @Summary 更新节点状态
// @Description 更新节点在线状态（管理员权限）
// @Tags 管理员-节点
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "节点ID"
// @Param request body UpdateNodeStatusRequest true "状态"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/nodes/{id}/status [put]
func (h *NodeHandler) UpdateNodeStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid node ID")
		return
	}

	var req UpdateNodeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters")
		return
	}

	if err := h.nodeService.UpdateNodeStatus(uint(id), req.Status); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Status changed, push updated config to the node.
	if NotifyNodeConfigChange != nil {
		node, err := h.nodeService.GetNodeByID(uint(id))
		if err == nil && node != nil {
			go NotifyNodeConfigChange(uint(id), node.Name, node.Type, node.Address, node.Port, node.ServerInfo, node.Rate, node.GroupID, node.ParentID)
		}
	}

	response.Success(c, nil)
}

// GlobalNodeMetricsFunc is set by cmd/server/main.go to bridge handler to grpc metrics cache.
// Returns the cached metrics for a node, or nil if not available.
var GlobalNodeMetricsFunc func(nodeID uint32) interface{}

// GetNodeMetrics 获取节点实时指标（从缓存中读取）
func (h *NodeHandler) GetNodeMetrics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid node ID")
		return
	}

	if GlobalNodeMetricsFunc == nil {
		response.InternalError(c, "Metrics cache not available")
		return
	}

	metrics := GlobalNodeMetricsFunc(uint32(id))
	if metrics == nil {
		response.NotFound(c, "No metrics available for this node")
		return
	}

	response.Success(c, metrics)
}

// CopyNode 复制节点
// @Summary 复制节点
// @Description 复制指定节点创建副本（管理员权限）
// @Tags 管理员-节点
// @Produce json
// @Security Bearer
// @Param id path int true "节点ID"
// @Success 200 {object} response.Response{data=model.Node}
// @Router /api/v1/admin/nodes/{id}/copy [post]
func (h *NodeHandler) CopyNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid node ID")
		return
	}

	node, err := h.nodeService.CopyNode(uint(id))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, node)
}

// ResetNodeTraffic 重置节点流量
// @Summary 重置节点流量
// @Description 重置指定节点的流量统计（管理员权限）
// @Tags 管理员-节点
// @Produce json
// @Security Bearer
// @Param id path int true "节点ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/nodes/{id}/reset-traffic [post]
func (h *NodeHandler) ResetNodeTraffic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid node ID")
		return
	}

	if err := h.nodeService.ResetNodeTraffic(uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// BatchResetNodeTrafficRequest 批量重置节点流量请求
type BatchResetNodeTrafficRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// BatchResetNodeTraffic 批量重置节点流量
// @Summary 批量重置节点流量
// @Description 批量重置多个节点的流量统计（管理员权限）
// @Tags 管理员-节点
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body BatchResetNodeTrafficRequest true "节点ID列表"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/nodes/batch-reset-traffic [post]
func (h *NodeHandler) BatchResetNodeTraffic(c *gin.Context) {
	var req BatchResetNodeTrafficRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if len(req.IDs) == 0 {
		response.BadRequest(c, "No node IDs provided")
		return
	}

	if err := h.nodeService.BatchResetNodeTraffic(req.IDs); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// BatchNodeRequest 批量节点操作请求
type BatchNodeRequest struct {
	IDs     []uint `json:"ids" binding:"required"`
	Action  string `json:"action" binding:"required,oneof=enable disable maintenance delete move"`
	GroupID *uint  `json:"group_id"`
}

// BatchNodes 批量节点操作
func (h *NodeHandler) BatchNodes(c *gin.Context) {
	var req BatchNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if len(req.IDs) == 0 {
		response.BadRequest(c, "No node IDs provided")
		return
	}

	var affected int64
	var err error

	switch req.Action {
	case "enable":
		err = h.nodeService.BatchUpdateStatus(req.IDs, 1)
	case "disable":
		err = h.nodeService.BatchUpdateStatus(req.IDs, 0)
	case "maintenance":
		err = h.nodeService.BatchUpdateStatus(req.IDs, 2)
	case "delete":
		err = h.nodeService.BatchDelete(req.IDs)
	case "move":
		if req.GroupID == nil {
			response.BadRequest(c, "group_id is required for move action")
			return
		}
		err = h.nodeService.BatchMoveGroup(req.IDs, *req.GroupID)
	default:
		response.BadRequest(c, "Invalid action")
		return
	}

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	affected = int64(len(req.IDs))
	response.Success(c, gin.H{"affected": affected})
}
