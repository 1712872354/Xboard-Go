package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// DeviceHandler 设备状态处理器
type DeviceHandler struct {
	deviceService service.DeviceService
}

// NewDeviceHandler 创建设备状态处理器
func NewDeviceHandler(deviceService service.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

// GetUserDevices 获取用户在线设备列表（管理员）
// @Summary 获取用户在线设备
// @Description 获取指定用户的在线设备列表（管理员权限）
// @Tags 管理员-设备
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=[]service.DeviceStatus}
// @Router /api/v1/admin/devices/user/{id} [get]
func (h *DeviceHandler) GetUserDevices(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	devices, err := h.deviceService.GetUserDevices(uint(userID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, devices)
}

// GetNodeDevices 获取节点在线设备列表（管理员）
// @Summary 获取节点在线设备
// @Description 获取指定节点的在线设备列表（管理员权限）
// @Tags 管理员-设备
// @Produce json
// @Security Bearer
// @Param id path int true "节点ID"
// @Success 200 {object} response.Response{data=[]service.DeviceStatus}
// @Router /api/v1/admin/devices/node/{id} [get]
func (h *DeviceHandler) GetNodeDevices(c *gin.Context) {
	nodeIDStr := c.Param("id")
	nodeID, err := strconv.ParseUint(nodeIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid node ID")
		return
	}

	devices, err := h.deviceService.GetNodeDevices(uint(nodeID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, devices)
}

// GetOnlineDeviceCount 获取用户在线设备数量（管理员）
// @Summary 获取在线设备数量
// @Description 获取指定用户的在线设备数量（管理员权限）
// @Tags 管理员-设备
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=int64}
// @Router /api/v1/admin/devices/user/{id}/count [get]
func (h *DeviceHandler) GetOnlineDeviceCount(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	count, err := h.deviceService.GetDeviceCount(uint(userID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"count": count})
}

// CleanupOfflineDevices 清理离线设备（管理员）
// @Summary 清理离线设备
// @Description 清理所有已过期的离线设备记录（管理员权限）
// @Tags 管理员-设备
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/devices/cleanup [post]
func (h *DeviceHandler) CleanupOfflineDevices(c *gin.Context) {
	if err := h.deviceService.CleanupOfflineDevices(); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
