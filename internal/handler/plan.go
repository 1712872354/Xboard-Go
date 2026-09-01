package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/model"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// PlanHandler 套餐处理器
type PlanHandler struct {
	planService service.PlanService
}

// NewPlanHandler 创建套餐处理器
func NewPlanHandler(planService service.PlanService) *PlanHandler {
	return &PlanHandler{
		planService: planService,
	}
}

// CreatePlanRequest 创建套餐请求
type CreatePlanRequest struct {
	Name         string  `json:"name" binding:"required"`
	Price        float64 `json:"price" binding:"required,min=0"`
	Traffic      int64   `json:"traffic" binding:"required,min=0"` // 流量（字节）
	DurationDays int     `json:"duration_days" binding:"required,min=1"`
	DeviceLimit  int     `json:"device_limit"`
	NodeGroup    string  `json:"node_group"`
	Description  string  `json:"description"`
}

// CreatePlan 创建套餐
// @Summary 创建套餐
// @Description 创建新的套餐（管理员权限）
// @Tags 管理员-套餐
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreatePlanRequest true "套餐信息"
// @Success 200 {object} response.Response{data=model.Plan}
// @Router /api/v1/admin/plans [post]
func (h *PlanHandler) CreatePlan(c *gin.Context) {
	var req CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	plan, err := h.planService.CreatePlan(
		req.Name,
		req.Price,
		req.Traffic,
		req.DurationDays,
		req.DeviceLimit,
		req.NodeGroup,
		req.Description,
	)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, plan)
}

// GetPlan 获取套餐详情
// @Summary 获取套餐详情
// @Description 获取指定套餐的详细信息
// @Tags 套餐
// @Produce json
// @Param id path int true "套餐ID"
// @Success 200 {object} response.Response{data=model.Plan}
// @Router /api/v1/plans/{id} [get]
func (h *PlanHandler) GetPlan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	plan, err := h.planService.GetPlanByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, plan)
}

// UpdatePlanRequest 更新套餐请求
type UpdatePlanRequest struct {
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Traffic      int64   `json:"traffic"`
	DurationDays int     `json:"duration_days"`
	DeviceLimit  int     `json:"device_limit"`
	NodeGroup    string  `json:"node_group"`
	Description  string  `json:"description"`
	Status       int     `json:"status"`
}

// UpdatePlan 更新套餐
// @Summary 更新套餐
// @Description 更新套餐信息（管理员权限）
// @Tags 管理员-套餐
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "套餐ID"
// @Param request body UpdatePlanRequest true "更新信息"
// @Success 200 {object} response.Response{data=model.Plan}
// @Router /api/v1/admin/plans/{id} [put]
func (h *PlanHandler) UpdatePlan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	var req UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	plan, err := h.planService.UpdatePlan(
		uint(id),
		req.Name,
		req.Price,
		req.Traffic,
		req.DurationDays,
		req.DeviceLimit,
		req.NodeGroup,
		req.Description,
		req.Status,
	)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, plan)
}

// DeletePlan 删除套餐
// @Summary 删除套餐
// @Description 删除套餐（管理员权限）
// @Tags 管理员-套餐
// @Produce json
// @Security Bearer
// @Param id path int true "套餐ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/plans/{id} [delete]
func (h *PlanHandler) DeletePlan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	if err := h.planService.DeletePlan(uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ListPlans 套餐列表（用户可见）
// @Summary 套餐列表
// @Description 获取上架的套餐列表（用户可见）
// @Tags 套餐
// @Produce json
// @Success 200 {object} response.Response{data=[]model.Plan}
// @Router /api/v1/plans [get]
func (h *PlanHandler) ListPlans(c *gin.Context) {
	plans, err := h.planService.ListActivePlans()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, plans)
}

// ListAllPlans 所有套餐列表（管理员）
// @Summary 所有套餐列表
// @Description 获取所有套餐列表（管理员权限）
// @Tags 管理员-套餐
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param include_disabled query bool false "是否包含已下架套餐" default(false)
// @Success 200 {object} response.Response
// @Router /api/v1/admin/plans [get]
func (h *PlanHandler) ListAllPlans(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	includeDisabled, _ := strconv.ParseBool(c.DefaultQuery("include_disabled", "false"))

	plans, total, err := h.planService.ListPlans(page, pageSize, includeDisabled)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      plans,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// OrderHandler 订单处理器
type OrderHandler struct {
	orderService service.OrderService
}

// NewOrderHandler 创建订单处理器
func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	PlanID     uint   `json:"plan_id" binding:"required"`
	CouponCode string `json:"coupon_code"`
}

// CreateOrder 创建订单
// @Summary 创建订单
// @Description 用户购买套餐，创建订单
// @Tags 订单
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateOrderRequest true "订单信息"
// @Success 200 {object} response.Response{data=model.Order}
// @Router /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := getCurrentUserID(c)

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	order, err := h.orderService.CreateOrder(userID, req.PlanID, req.CouponCode)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, order)
}

// GetOrder 获取订单详情
// @Summary 获取订单详情
// @Description 获取指定订单的详细信息
// @Tags 订单
// @Produce json
// @Security Bearer
// @Param id path int true "订单ID"
// @Success 200 {object} response.Response{data=model.Order}
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	order, err := h.orderService.GetOrderByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, order)
}

// ListOrders 用户订单列表
// @Summary 我的订单
// @Description 获取当前用户的订单列表
// @Tags 订单
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response
// @Router /api/v1/orders [get]
func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID := getCurrentUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	orders, total, err := h.orderService.ListUserOrders(userID, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CancelOrder 取消订单
// @Summary 取消订单
// @Description 用户取消待支付的订单
// @Tags 订单
// @Produce json
// @Security Bearer
// @Param id path int true "订单ID"
// @Success 200 {object} response.Response
// @Router /api/v1/orders/{id}/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID := getCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	if err := h.orderService.CancelOrder(uint(id), userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ListAllOrders 所有订单列表（管理员）
// @Summary 所有订单列表
// @Description 获取所有订单列表（管理员权限）
// @Tags 管理员-订单
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query int false "订单状态筛选 (-1为全部)" default(-1)
// @Param user_id query int false "用户ID筛选" default(0)
// @Success 200 {object} response.Response
// @Router /api/v1/admin/orders [get]
func (h *OrderHandler) ListAllOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 32)

	orders, total, err := h.orderService.ListOrders(page, pageSize, status, uint(userID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ConfirmPaymentRequest 确认支付请求
type ConfirmPaymentRequest struct {
	TradeNo string `json:"trade_no" binding:"required"`
}

// ConfirmPayment 确认支付（模拟支付回调）
// @Summary 确认支付
// @Description 手动确认订单支付（模拟支付回调，管理员权限）
// @Tags 管理员-订单
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body ConfirmPaymentRequest true "支付信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/orders/confirm-payment [post]
func (h *OrderHandler) ConfirmPayment(c *gin.Context) {
	var req ConfirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if err := h.orderService.ConfirmPayment(req.TradeNo); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// AssignOrderRequest 分配订单请求
type AssignOrderRequest struct {
	Email  string  `json:"email" binding:"required,email"`
	PlanID uint    `json:"plan_id" binding:"required"`
	Period string  `json:"period"`
	Amount float64 `json:"amount"`
}

// AssignOrder 管理员为用户分配订单
func (h *OrderHandler) AssignOrder(c *gin.Context) {
	var req AssignOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	order, err := h.orderService.AssignOrder(req.Email, req.PlanID, req.Period, req.Amount)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, order)
}

// getCurrentUserID 从 context 获取当前用户ID
// 注意：使用前需确保已通过 JWT 中间件
func getCurrentUserID(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(uint)
}

// 确保 model 包被引用（避免未使用导入错误）
var _ = model.User{}
