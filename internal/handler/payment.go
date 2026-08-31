package handler

import (
	"fmt"
	"html"
	"net/http"

	"xboard-go/internal/middleware"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	paymentService service.PaymentService
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	OrderID       uint   `json:"order_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

// CreatePayment 创建支付
// @Summary 创建支付
// @Description 为订单创建支付链接
// @Tags 支付
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreatePaymentRequest true "支付信息"
// @Success 200 {object} response.Response{data=service.PaymentResult}
// @Router /api/v1/payment/create [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	result, err := h.paymentService.CreatePayment(req.OrderID, userID, req.PaymentMethod)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// ListMethods 获取支付方式列表
// @Summary 获取支付方式
// @Description 获取可用的支付方式列表
// @Tags 支付
// @Produce json
// @Success 200 {object} response.Response{data=[]service.PaymentMethodInfo}
// @Router /api/v1/payment/methods [get]
func (h *PaymentHandler) ListMethods(c *gin.Context) {
	methods := h.paymentService.ListMethods()
	response.Success(c, methods)
}

// MockPay 模拟支付页面
// @Summary 模拟支付
// @Description 模拟支付跳转页面（仅用于测试）
// @Tags 支付
// @Produce html
// @Param trade_no query string true "商户订单号"
// @Param amount query string true "支付金额"
// @Router /api/v1/payment/mock/pay [get]
func (h *PaymentHandler) MockPay(c *gin.Context) {
	tradeNo := html.EscapeString(c.Query("trade_no"))
	amount := html.EscapeString(c.Query("amount"))

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><title>模拟支付</title></head>
<body style="text-align:center; padding: 50px; font-family: sans-serif;">
    <h2>模拟支付页面</h2>
    <p>订单号: %s</p>
    <p>金额: ¥%s</p>
    <p style="color: #666;">这是一个模拟支付页面，实际环境中这里是支付宝/微信支付页面。</p>
    <p style="color: #999;">请使用管理员接口确认支付，或调用回调接口。</p>
</body>
</html>`, tradeNo, amount)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, htmlContent)
}

// MockCallback 模拟支付回调
// @Summary 模拟支付回调
// @Description 模拟支付成功后的回调（仅用于测试）
// @Tags 支付
// @Produce plain
// @Param trade_no query string true "商户订单号"
// @Param amount query string false "支付金额" default("0.01")
// @Success 200 {string} string "success"
// @Router /api/v1/payment/mock/callback [get]
func (h *PaymentHandler) MockCallback(c *gin.Context) {
	tradeNo := c.Query("trade_no")
	if tradeNo == "" {
		c.String(http.StatusBadRequest, "trade_no is required")
		return
	}

	amountStr := c.DefaultQuery("amount", "0.01")
	var amount float64
	fmt.Sscanf(amountStr, "%f", &amount)

	// 使用支付服务生成带正确签名的回调参数
	params := h.paymentService.GenerateMockCallbackParams(tradeNo, amount)

	result, err := h.paymentService.HandleCallback("mock", params)
	if err != nil {
		c.String(http.StatusInternalServerError, "error: %v", err)
		return
	}

	c.String(http.StatusOK, result)
}

// RedeemHandler 兑换码处理器
type RedeemHandler struct {
	redeemService service.RedeemService
}

// NewRedeemHandler 创建兑换码处理器
func NewRedeemHandler(redeemService service.RedeemService) *RedeemHandler {
	return &RedeemHandler{
		redeemService: redeemService,
	}
}

// RedeemRequest 兑换请求
type RedeemRequest struct {
	Code string `json:"code" binding:"required"`
}

// Redeem 兑换码兑换
// @Summary 兑换卡密
// @Description 使用兑换码兑换套餐
// @Tags 卡密
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body RedeemRequest true "兑换码"
// @Success 200 {object} response.Response
// @Router /api/v1/redeem [post]
func (h *RedeemHandler) Redeem(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	plan, err := h.redeemService.Redeem(req.Code, userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "兑换成功",
		"plan":    plan,
	})
}

// GenerateRedeemRequest 生成兑换码请求
type GenerateRedeemRequest struct {
	PlanID uint   `json:"plan_id" binding:"required"`
	Count  int    `json:"count" binding:"required,min=1,max=1000"`
	Prefix string `json:"prefix"`
}

// Generate 批量生成兑换码
// @Summary 生成兑换码
// @Description 批量生成兑换码（管理员权限）
// @Tags 管理员-卡密
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body GenerateRedeemRequest true "生成信息"
// @Success 200 {object} response.Response{data=[]model.RedeemCode}
// @Router /api/v1/admin/redeem/generate [post]
func (h *RedeemHandler) Generate(c *gin.Context) {
	var req GenerateRedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	codes, err := h.redeemService.Generate(req.PlanID, req.Count, req.Prefix)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":  codes,
		"count": len(codes),
	})
}

// List 兑换码列表
// @Summary 兑换码列表
// @Description 获取兑换码列表（管理员权限）
// @Tags 管理员-卡密
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query int false "状态筛选 (-1=全部, 0=未使用, 1=已使用)" default(-1)
// @Param plan_id query int false "套餐ID筛选" default(0)
// @Success 200 {object} response.Response
// @Router /api/v1/admin/redeem [get]
func (h *RedeemHandler) List(c *gin.Context) {
	page := strconvAtoi(c.DefaultQuery("page", "1"))
	pageSize := strconvAtoi(c.DefaultQuery("page_size", "20"))
	status := strconvAtoi(c.DefaultQuery("status", "-1"))
	planID := strconvAtoi(c.DefaultQuery("plan_id", "0"))

	codes, total, err := h.redeemService.List(page, pageSize, status, uint(planID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      codes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Delete 删除兑换码
// @Summary 删除兑换码
// @Description 删除未使用的兑换码（管理员权限）
// @Tags 管理员-卡密
// @Produce json
// @Security Bearer
// @Param id path int true "兑换码ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/redeem/{id} [delete]
func (h *RedeemHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconvParseUint(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid redeem code ID")
		return
	}

	if err := h.redeemService.Delete(uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetStats 获取兑换码统计
// @Summary 兑换码统计
// @Description 获取兑换码使用统计（管理员权限）
// @Tags 管理员-卡密
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=service.RedeemStats}
// @Router /api/v1/admin/redeem/stats [get]
func (h *RedeemHandler) GetStats(c *gin.Context) {
	stats, err := h.redeemService.GetStats()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

// 辅助函数
func strconvAtoi(s string) int {
	n := 0
	fmt.Sscanf(s, "%d", &n)
	return n
}

func strconvParseUint(s string) (uint, error) {
	var n uint
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
