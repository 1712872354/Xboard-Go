package handler

import (
	"xboard-go/internal/middleware"
	"xboard-go/internal/model"
	"xboard-go/internal/service"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// OrderCheckoutHandler 订单结账处理器
type OrderCheckoutHandler struct {
	orderService   service.OrderService
	paymentService service.PaymentService
}

// NewOrderCheckoutHandler 创建订单结账处理器
func NewOrderCheckoutHandler(orderService service.OrderService, paymentService service.PaymentService) *OrderCheckoutHandler {
	return &OrderCheckoutHandler{
		orderService:   orderService,
		paymentService: paymentService,
	}
}

// CheckoutRequest 结账请求
type CheckoutRequest struct {
	TradeNo string `json:"trade_no" binding:"required"`
	Method  uint   `json:"method" binding:"required"`
}

// Checkout 订单结账
// @Summary 订单结账
// @Description 用户选择支付方式进行结账
// @Tags 订单
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CheckoutRequest true "结账信息"
// @Success 200 {object} response.Response
// @Router /api/v1/orders/checkout [post]
func (h *OrderCheckoutHandler) Checkout(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	// 获取订单
	order, err := h.orderService.GetOrderByTradeNo(req.TradeNo)
	if err != nil {
		response.BadRequest(c, "订单不存在")
		return
	}

	// 验证订单归属
	if order.UserID != userID {
		response.BadRequest(c, "无权操作此订单")
		return
	}

	// 验证订单状态
	if !order.IsPending() {
		response.BadRequest(c, "订单状态异常")
		return
	}

	// 免费订单直接完成
	if order.ActualAmount <= 0 {
		if err := h.orderService.ConfirmPayment(order.TradeNo); err != nil {
			response.InternalError(c, "支付失败: "+err.Error())
			return
		}
		response.Success(c, gin.H{
			"type": -1,
			"data": true,
		})
		return
	}

	// 获取支付方式
	db := database.Get()
	var payment model.Payment
	if err := db.Where("id = ? AND status = ?", req.Method, 1).First(&payment).Error; err != nil {
		response.BadRequest(c, "支付方式不可用")
		return
	}

	// 更新订单支付方式
	order.PaymentMethod = payment.Payment
	if err := db.Save(order).Error; err != nil {
		response.InternalError(c, "更新订单失败")
		return
	}

	// 创建支付
	result, err := h.paymentService.CreatePayment(order.ID, userID, payment.Payment)
	if err != nil {
		response.BadRequest(c, "创建支付失败: "+err.Error())
		return
	}

	response.Success(c, result)
}

// CheckOrder 检查订单支付状态
// @Summary 检查订单状态
// @Description 检查订单是否已支付
// @Tags 订单
// @Produce json
// @Security Bearer
// @Param trade_no query string true "商户订单号"
// @Success 200 {object} response.Response
// @Router /api/v1/orders/check [get]
func (h *OrderCheckoutHandler) CheckOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tradeNo := c.Query("trade_no")

	if tradeNo == "" {
		response.BadRequest(c, "订单号不能为空")
		return
	}

	order, err := h.orderService.GetOrderByTradeNo(tradeNo)
	if err != nil {
		response.BadRequest(c, "订单不存在")
		return
	}

	if order.UserID != userID {
		response.BadRequest(c, "无权查看此订单")
		return
	}

	response.Success(c, gin.H{
		"trade_no": order.TradeNo,
		"status":   order.Status,
		"is_paid":  order.IsPaid(),
	})
}

// GetOrderDetail 获取订单详情（用户端）
// @Summary 订单详情
// @Description 获取订单详细信息
// @Tags 订单
// @Produce json
// @Security Bearer
// @Param trade_no query string true "商户订单号"
// @Success 200 {object} response.Response
// @Router /api/v1/orders/detail [get]
func (h *OrderCheckoutHandler) GetOrderDetail(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tradeNo := c.Query("trade_no")

	if tradeNo == "" {
		response.BadRequest(c, "订单号不能为空")
		return
	}

	order, err := h.orderService.GetOrderByTradeNo(tradeNo)
	if err != nil {
		response.BadRequest(c, "订单不存在")
		return
	}

	if order.UserID != userID {
		response.BadRequest(c, "无权查看此订单")
		return
	}

	// 加载关联数据
	db := database.Get()
	db.Preload("Plan").First(order, order.ID)

	// 获取可用支付方式
	var payments []model.Payment
	db.Where("status = ?", 1).Order("sort ASC").Find(&payments)

	response.Success(c, gin.H{
		"order":    order,
		"payments": payments,
	})
}

// GetUserPaymentMethods 获取用户可用的支付方式
// @Summary 获取支付方式
// @Description 获取用户可用的支付方式列表
// @Tags 订单
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/orders/payment-methods [get]
func (h *OrderCheckoutHandler) GetUserPaymentMethods(c *gin.Context) {
	db := database.Get()

	var payments []model.Payment
	if err := db.Where("status = ?", 1).Order("sort ASC").Find(&payments).Error; err != nil {
		response.InternalError(c, "获取支付方式失败")
		return
	}

	// 转换为简化格式
	var methods []gin.H
	for _, p := range payments {
		methods = append(methods, gin.H{
			"id":      p.ID,
			"name":    p.Name,
			"payment": p.Payment,
			"icon":    p.Icon,
		})
	}

	if methods == nil {
		methods = []gin.H{}
	}

	response.Success(c, methods)
}
