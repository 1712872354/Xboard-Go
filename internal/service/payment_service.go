package service

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"xboard-go/internal/model"
)

// PaymentGateway 支付网关接口
// 所有支付渠道（支付宝、微信、模拟支付等）都实现此接口
type PaymentGateway interface {
	// Name 支付渠道名称
	Name() string
	// CreatePayment 创建支付订单
	CreatePayment(order *model.Order) (*PaymentResult, error)
	// VerifyCallback 验证支付回调签名
	VerifyCallback(params map[string]string) (bool, *CallbackResult, error)
	// GetTradeNo 从回调参数中获取商户订单号
	GetTradeNo(params map[string]string) string
}

// PaymentResult 支付创建结果
type PaymentResult struct {
	PaymentURL  string            `json:"payment_url"`  // 支付跳转链接（或二维码内容）
	TradeNo     string            `json:"trade_no"`     // 商户订单号
	Amount      float64           `json:"amount"`       // 支付金额
	PaymentID   string            `json:"payment_id"`   // 第三方支付单号
	ExpireAt    time.Time         `json:"expire_at"`    // 过期时间
	Extra       map[string]string `json:"extra,omitempty"` // 额外信息
}

// CallbackResult 支付回调结果
type CallbackResult struct {
	TradeNo    string  `json:"trade_no"`     // 商户订单号
	PaymentID  string  `json:"payment_id"`   // 第三方支付单号
	Amount     float64 `json:"amount"`       // 支付金额
	Status     string  `json:"status"`       // 支付状态：success/failed
	PayTime    string  `json:"pay_time"`     // 支付时间
	RawParams  map[string]string `json:"-"`  // 原始参数（不返回给前端）
}

// ===== 模拟支付网关 =====

// MockPaymentGateway 模拟支付网关
// 用于开发测试，直接返回模拟的支付链接
type MockPaymentGateway struct {
	baseURL string
}

// NewMockPaymentGateway 创建模拟支付网关
func NewMockPaymentGateway(baseURL string) PaymentGateway {
	return &MockPaymentGateway{
		baseURL: baseURL,
	}
}

func (g *MockPaymentGateway) Name() string {
	return "mock"
}

func (g *MockPaymentGateway) CreatePayment(order *model.Order) (*PaymentResult, error) {
	if order == nil {
		return nil, errors.New("order is nil")
	}

	// 模拟支付链接：实际生产中这里是支付宝/微信的支付地址
	paymentURL := fmt.Sprintf("%s/api/v1/payment/mock/pay?trade_no=%s&amount=%.2f",
		g.baseURL, order.TradeNo, order.Amount)

	return &PaymentResult{
		PaymentURL: paymentURL,
		TradeNo:    order.TradeNo,
		Amount:     order.Amount,
		PaymentID:  fmt.Sprintf("MOCK%s", order.TradeNo),
		ExpireAt:   time.Now().Add(30 * time.Minute),
		Extra: map[string]string{
			"method": "mock",
			"tip":    "This is a mock payment. Use admin API to confirm payment.",
		},
	}, nil
}

func (g *MockPaymentGateway) VerifyCallback(params map[string]string) (bool, *CallbackResult, error) {
	if params == nil {
		return false, nil, errors.New("params is nil")
	}

	// 模拟支付的签名验证
	sign := params["sign"]
	expectedSign := g.generateMockSign(params)

	if sign != expectedSign {
		return false, nil, errors.New("invalid signature")
	}

	status := params["status"]
	if status == "" {
		status = "success"
	}

	result := &CallbackResult{
		TradeNo:   params["trade_no"],
		PaymentID: params["payment_id"],
		Amount:    parseFloat(params["amount"]),
		Status:    status,
		PayTime:   params["pay_time"],
		RawParams: params,
	}

	return true, result, nil
}

func (g *MockPaymentGateway) GetTradeNo(params map[string]string) string {
	return params["trade_no"]
}

// generateMockSign 生成模拟签名
func (g *MockPaymentGateway) generateMockSign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && k != "sign_type" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	signStr := strings.Join(parts, "&") + "&key=mock_payment_secret_key"

	h := md5.Sum([]byte(signStr))
	return hex.EncodeToString(h[:])
}

// ===== 支付管理服务 =====

// PaymentService 支付服务
type PaymentService interface {
	// GetGateway 获取指定支付渠道的网关
	GetGateway(method string) (PaymentGateway, error)
	// CreatePayment 创建支付
	CreatePayment(orderID uint, userID uint, method string) (*PaymentResult, error)
	// HandleCallback 处理支付回调
	HandleCallback(method string, params map[string]string) (string, error)
	// ListMethods 获取可用支付方式列表
	ListMethods() []PaymentMethodInfo
	// GenerateMockCallbackParams 生成模拟支付回调参数（带签名，仅测试用）
	GenerateMockCallbackParams(tradeNo string, amount float64) map[string]string
}

type paymentService struct {
	gateways    map[string]PaymentGateway
	orderRepo   orderRepoInterface
	userService UserService
}

// PaymentMethodInfo 支付方式信息
type PaymentMethodInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	Status int    `json:"status"` // 1: 启用, 0: 禁用
}

// orderRepoInterface 订单仓储接口（避免循环依赖）
type orderRepoInterface interface {
	GetByID(id uint) (*model.Order, error)
	GetByTradeNo(tradeNo string) (*model.Order, error)
	UpdateStatus(orderID uint, status int, paidAt *time.Time) error
}

// NewPaymentService 创建支付服务
func NewPaymentService(orderRepo orderRepoInterface, userService UserService, baseURL string) PaymentService {
	gateways := make(map[string]PaymentGateway)

	// 注册模拟支付网关
	mockGateway := NewMockPaymentGateway(baseURL)
	gateways[mockGateway.Name()] = mockGateway

	return &paymentService{
		gateways:    gateways,
		orderRepo:   orderRepo,
		userService: userService,
	}
}

func (s *paymentService) GetGateway(method string) (PaymentGateway, error) {
	gateway, exists := s.gateways[method]
	if !exists {
		return nil, fmt.Errorf("unsupported payment method: %s", method)
	}
	return gateway, nil
}

func (s *paymentService) CreatePayment(orderID uint, userID uint, method string) (*PaymentResult, error) {
	// 获取订单
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	// 验证订单归属
	if order.UserID != userID {
		return nil, errors.New("order does not belong to user")
	}

	// 检查订单状态
	if order.IsPaid() {
		return nil, errors.New("order already paid")
	}
	if !order.IsPending() {
		return nil, errors.New("order is not pending")
	}

	// 获取支付网关
	gateway, err := s.GetGateway(method)
	if err != nil {
		return nil, err
	}

	// 更新订单支付方式
	order.PaymentMethod = method
	// 注意：这里不直接更新订单，由调用方决定是否保存

	// 创建支付
	result, err := gateway.CreatePayment(order)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return result, nil
}

func (s *paymentService) HandleCallback(method string, params map[string]string) (string, error) {
	gateway, err := s.GetGateway(method)
	if err != nil {
		return "", err
	}

	// 验证签名
	valid, callbackResult, err := gateway.VerifyCallback(params)
	if err != nil {
		return "", fmt.Errorf("failed to verify callback: %w", err)
	}
	if !valid {
		return "", errors.New("invalid callback signature")
	}

	// 获取订单
	tradeNo := gateway.GetTradeNo(params)
	order, err := s.orderRepo.GetByTradeNo(tradeNo)
	if err != nil {
		return "", fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return "", errors.New("order not found")
	}

	// 如果已支付，直接返回成功（幂等）
	if order.IsPaid() {
		return "success", nil
	}

	// 处理支付结果
	if callbackResult.Status == "success" {
		// 更新订单状态
		now := time.Now()
		if err := s.orderRepo.UpdateStatus(order.ID, model.OrderStatusPaid, &now); err != nil {
			return "", fmt.Errorf("failed to update order status: %w", err)
		}

		// 为用户增加流量和时长
		if err := s.userService.AddTraffic(order.UserID, order.Plan.Traffic, order.Plan.DurationDays); err != nil {
			return "", fmt.Errorf("failed to add traffic: %w", err)
		}
	}

	return "success", nil
}

func (s *paymentService) ListMethods() []PaymentMethodInfo {
	var methods []PaymentMethodInfo
	for id := range s.gateways {
		methods = append(methods, PaymentMethodInfo{
			ID:     id,
			Name:   getPaymentMethodName(id),
			Status: 1,
		})
	}
	return methods
}

// GenerateMockCallbackParams 生成模拟支付回调参数（带有效签名）
// 仅用于开发/测试环境
func (s *paymentService) GenerateMockCallbackParams(tradeNo string, amount float64) map[string]string {
	params := map[string]string{
		"trade_no":   tradeNo,
		"payment_id": fmt.Sprintf("MOCK%s", tradeNo),
		"amount":     fmt.Sprintf("%.2f", amount),
		"status":     "success",
		"pay_time":   time.Now().Format("2006-01-02 15:04:05"),
	}

	// 生成签名（与 MockPaymentGateway 使用相同算法）
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && k != "sign_type" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	signStr := strings.Join(parts, "&") + "&key=mock_payment_secret_key"

	h := md5.Sum([]byte(signStr))
	params["sign"] = hex.EncodeToString(h[:])

	return params
}

func getPaymentMethodName(method string) string {
	switch method {
	case "mock":
		return "模拟支付"
	case "alipay":
		return "支付宝"
	case "wechat":
		return "微信支付"
	default:
		return method
	}
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
