package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// ServerMachineHandler 服务器机器处理器
type ServerMachineHandler struct {
	machineService service.ServerMachineService
}

// NewServerMachineHandler 创建服务器机器处理器
func NewServerMachineHandler(machineService service.ServerMachineService) *ServerMachineHandler {
	return &ServerMachineHandler{
		machineService: machineService,
	}
}

// CreateServerMachineRequest 创建服务器机器请求
type CreateServerMachineRequest struct {
	Name     string `json:"name" binding:"required"`
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	Protocol string `json:"protocol"`
}

// CreateServerMachine 创建服务器机器（管理员）
func (h *ServerMachineHandler) CreateServerMachine(c *gin.Context) {
	var req CreateServerMachineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	machine, err := h.machineService.Create(req.Name, req.Host, req.Port, req.Protocol)
	if err != nil {
		response.InternalError(c, "创建服务器机器失败："+err.Error())
		return
	}

	response.Success(c, machine)
}

// GetServerMachine 获取服务器机器详情（管理员）
func (h *ServerMachineHandler) GetServerMachine(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的机器ID")
		return
	}

	machine, err := h.machineService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "服务器机器不存在")
		return
	}

	response.Success(c, machine)
}

// UpdateServerMachineRequest 更新服务器机器请求
type UpdateServerMachineRequest struct {
	Name     string `json:"name" binding:"required"`
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	Protocol string `json:"protocol"`
	Status   int    `json:"status"`
}

// UpdateServerMachine 更新服务器机器（管理员）
func (h *ServerMachineHandler) UpdateServerMachine(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的机器ID")
		return
	}

	var req UpdateServerMachineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	machine, err := h.machineService.Update(uint(id), req.Name, req.Host, req.Port, req.Protocol, req.Status)
	if err != nil {
		response.InternalError(c, "更新服务器机器失败："+err.Error())
		return
	}

	response.Success(c, machine)
}

// DeleteServerMachine 删除服务器机器（管理员）
func (h *ServerMachineHandler) DeleteServerMachine(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的机器ID")
		return
	}

	if err := h.machineService.Delete(uint(id)); err != nil {
		response.InternalError(c, "删除服务器机器失败")
		return
	}

	response.Success(c, nil)
}

// ListServerMachines 服务器机器列表（管理员）
func (h *ServerMachineHandler) ListServerMachines(c *gin.Context) {
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

	machines, total, err := h.machineService.List(page, pageSize)
	if err != nil {
		response.InternalError(c, "获取服务器机器列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  machines,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// ListAllServerMachines 获取所有服务器机器
func (h *ServerMachineHandler) ListAllServerMachines(c *gin.Context) {
	machines, err := h.machineService.ListAll()
	if err != nil {
		response.InternalError(c, "获取服务器机器列表失败")
		return
	}

	response.Success(c, machines)
}

// UpdateServerMachineStatusRequest 更新服务器机器状态请求
type UpdateServerMachineStatusRequest struct {
	Status int `json:"status"`
}

// UpdateServerMachineStatus 更新服务器机器状态（管理员）
func (h *ServerMachineHandler) UpdateServerMachineStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的机器ID")
		return
	}

	var req UpdateServerMachineStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.machineService.UpdateStatus(uint(id), req.Status); err != nil {
		response.InternalError(c, "更新服务器机器状态失败："+err.Error())
		return
	}

	response.Success(c, nil)
}

// UpdateServerMachineLoadRequest 更新服务器机器负载请求
type UpdateServerMachineLoadRequest struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Disk   float64 `json:"disk"`
}

// UpdateServerMachineLoad 更新服务器机器负载（管理员）
func (h *ServerMachineHandler) UpdateServerMachineLoad(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的机器ID")
		return
	}

	var req UpdateServerMachineLoadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.machineService.UpdateLoad(uint(id), req.CPU, req.Memory, req.Disk); err != nil {
		response.InternalError(c, "更新服务器机器负载失败："+err.Error())
		return
	}

	response.Success(c, nil)
}

// ResetToken 重置节点认证Token（管理员）
func (h *ServerMachineHandler) ResetToken(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的机器ID")
		return
	}

	machine, err := h.machineService.ResetToken(uint(id))
	if err != nil {
		response.InternalError(c, "重置Token失败："+err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":    machine.ID,
		"token": machine.Token,
	})
}

// GetInstallCommand 获取节点安装命令（管理员）
func (h *ServerMachineHandler) GetInstallCommand(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的机器ID")
		return
	}

	cmd, err := h.machineService.GetInstallCommand(uint(id))
	if err != nil {
		response.InternalError(c, "获取安装命令失败："+err.Error())
		return
	}

	response.Success(c, gin.H{
		"install_command": cmd,
	})
}
