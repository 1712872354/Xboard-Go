package handler

import (
	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// PaymentGatewayHandler 支付网关处理器
type PaymentGatewayHandler struct {
	gatewayService service.PaymentGatewayService
}

// NewPaymentGatewayHandler 创建支付网关处理器
func NewPaymentGatewayHandler(gatewayService service.PaymentGatewayService) *PaymentGatewayHandler {
	return &PaymentGatewayHandler{
		gatewayService: gatewayService,
	}
}

// CreateGatewayRequest 创建支付网关请求
type CreateGatewayRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Icon         string                 `json:"icon"`
	Payment      string                 `json:"payment" binding:"required"`
	Config       map[string]interface{} `json:"config"`
	NotifyDomain string                 `json:"notify_domain"`
}

// CreateGateway 创建支付网关
// @Summary 创建支付网关
// @Description 创建新的支付网关（管理员权限）
// @Tags 管理员-支付
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateGatewayRequest true "支付网关信息"
// @Success 200 {object} response.Response{data=model.Payment}
// @Router /api/v1/admin/payment/gateways [post]
func (h *PaymentGatewayHandler) CreateGateway(c *gin.Context) {
	var req CreateGatewayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	gateway, err := h.gatewayService.CreateGateway(req.Name, req.Icon, req.Payment, req.Config, req.NotifyDomain)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gateway)
}

// GetGateway 获取支付网关
// @Summary 获取支付网关详情
// @Description 获取指定支付网关的详细信息（管理员权限）
// @Tags 管理员-支付
// @Produce json
// @Security Bearer
// @Param id path int true "网关ID"
// @Success 200 {object} response.Response{data=model.Payment}
// @Router /api/v1/admin/payment/gateways/{id} [get]
func (h *PaymentGatewayHandler) GetGateway(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid gateway ID")
		return
	}

	gateway, err := h.gatewayService.GetGateway(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, gateway)
}

// UpdateGatewayRequest 更新支付网关请求
type UpdateGatewayRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Icon         string                 `json:"icon"`
	Payment      string                 `json:"payment" binding:"required"`
	Config       map[string]interface{} `json:"config"`
	NotifyDomain string                 `json:"notify_domain"`
}

// UpdateGateway 更新支付网关
// @Summary 更新支付网关
// @Description 更新指定支付网关的信息（管理员权限）
// @Tags 管理员-支付
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "网关ID"
// @Param request body UpdateGatewayRequest true "支付网关信息"
// @Success 200 {object} response.Response{data=model.Payment}
// @Router /api/v1/admin/payment/gateways/{id} [put]
func (h *PaymentGatewayHandler) UpdateGateway(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid gateway ID")
		return
	}

	var req UpdateGatewayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	gateway, err := h.gatewayService.UpdateGateway(id, req.Name, req.Icon, req.Payment, req.Config, req.NotifyDomain)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gateway)
}

// DeleteGateway 删除支付网关
// @Summary 删除支付网关
// @Description 删除指定支付网关（管理员权限）
// @Tags 管理员-支付
// @Produce json
// @Security Bearer
// @Param id path int true "网关ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/payment/gateways/{id} [delete]
func (h *PaymentGatewayHandler) DeleteGateway(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid gateway ID")
		return
	}

	if err := h.gatewayService.DeleteGateway(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ListGateways 获取支付网关列表
// @Summary 支付网关列表
// @Description 获取所有支付网关列表（管理员权限）
// @Tags 管理员-支付
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=[]model.Payment}
// @Router /api/v1/admin/payment/gateways [get]
func (h *PaymentGatewayHandler) ListGateways(c *gin.Context) {
	gateways, err := h.gatewayService.ListGateways()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gateways)
}

// UpdateGatewayStatusRequest 更新网关状态请求
type UpdateGatewayStatusRequest struct {
	Status int `json:"status" binding:"required,oneof=0 1"`
}

// UpdateGatewayStatus 更新支付网关状态
// @Summary 更新支付网关状态
// @Description 启用或禁用支付网关（管理员权限）
// @Tags 管理员-支付
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "网关ID"
// @Param request body UpdateGatewayStatusRequest true "状态"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/payment/gateways/{id}/status [put]
func (h *PaymentGatewayHandler) UpdateGatewayStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid gateway ID")
		return
	}

	var req UpdateGatewayStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters")
		return
	}

	if err := h.gatewayService.UpdateGatewayStatus(id, req.Status); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// UpdateGatewaySortRequest 更新网关排序请求
type UpdateGatewaySortRequest struct {
	Sort int `json:"sort" binding:"required"`
}

// UpdateGatewaySort 更新支付网关排序
// @Summary 更新支付网关排序
// @Description 更新支付网关的排序值（管理员权限）
// @Tags 管理员-支付
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "网关ID"
// @Param request body UpdateGatewaySortRequest true "排序"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/payment/gateways/{id}/sort [put]
func (h *PaymentGatewayHandler) UpdateGatewaySort(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid gateway ID")
		return
	}

	var req UpdateGatewaySortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters")
		return
	}

	if err := h.gatewayService.UpdateGatewaySort(id, req.Sort); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}
