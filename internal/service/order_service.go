package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
	"xboard-go/pkg/database"
	"xboard-go/pkg/utils"
)

// OrderService 订单服务接口
type OrderService interface {
	CreateOrder(userID, planID uint, couponCode string) (*model.Order, error)
	GetOrderByID(id uint) (*model.Order, error)
	GetOrderByTradeNo(tradeNo string) (*model.Order, error)
	ListUserOrders(userID uint, page, pageSize int) ([]model.Order, int64, error)
	ListOrders(page, pageSize int, status int, userID uint) ([]model.Order, int64, error)
	ConfirmPayment(tradeNo string) error
	CancelOrder(orderID uint, userID uint) error
	AssignOrder(email string, planID uint, period string, amount float64) (*model.Order, error)
}

type orderService struct {
	orderRepo  repository.OrderRepository
	planRepo   repository.PlanRepository
	userRepo   repository.UserRepository
	couponRepo repository.CouponRepository
}

// NewOrderService 创建订单服务
func NewOrderService(orderRepo repository.OrderRepository, planRepo repository.PlanRepository, userRepo repository.UserRepository, couponRepo repository.CouponRepository) OrderService {
	return &orderService{
		orderRepo:  orderRepo,
		planRepo:   planRepo,
		userRepo:   userRepo,
		couponRepo: couponRepo,
	}
}

// CreateOrder 创建订单
func (s *orderService) CreateOrder(userID, planID uint, couponCode string) (*model.Order, error) {
	// 检查用户是否存在
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 检查套餐是否存在且可用
	plan, err := s.planRepo.GetByID(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}
	if plan == nil {
		return nil, errors.New("plan not found")
	}
	if !plan.IsAvailable() {
		return nil, errors.New("plan is not available")
	}

	// 生成商户订单号
	tradeNo, err := generateTradeNo()
	if err != nil {
		return nil, fmt.Errorf("failed to generate trade no: %w", err)
	}

	// 计算折扣
	var discount float64
	if couponCode != "" && s.couponRepo != nil {
		coupon, err := s.couponRepo.GetByCode(couponCode)
		if err == nil && coupon != nil {
			if !coupon.IsActive() {
				return nil, errors.New("coupon is not active")
			}

			// 检查用户是否可用
			if coupon.UserIDs != "" {
				userIDStr := fmt.Sprintf("%d", userID)
				found := false
				for _, id := range strings.Split(coupon.UserIDs, ",") {
					if strings.TrimSpace(id) == userIDStr {
						found = true
						break
					}
				}
				if !found {
					return nil, errors.New("coupon not available for this user")
				}
			}

			// 检查套餐是否可用
			if coupon.PlanIDs != "" {
				planIDStr := fmt.Sprintf("%d", planID)
				found := false
				for _, id := range strings.Split(coupon.PlanIDs, ",") {
					if strings.TrimSpace(id) == planIDStr {
						found = true
						break
					}
				}
				if !found {
					return nil, errors.New("coupon not available for this plan")
				}
			}

			discount = coupon.CalculateDiscount(plan.Price)
		}
	}

	// 创建订单
	order := &model.Order{
		UserID:        userID,
		PlanID:        planID,
		Amount:        plan.Price,
		CouponCode:    couponCode,
		Discount:      discount,
		ActualAmount:  plan.Price - discount,
		Status:        model.OrderStatusPending,
		PaymentMethod: "manual",
		TradeNo:       tradeNo,
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 重新加载订单（包含套餐信息）
	return s.orderRepo.GetByID(order.ID)
}

// GetOrderByID 根据ID获取订单
func (s *orderService) GetOrderByID(id uint) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	return order, nil
}

// GetOrderByTradeNo 根据商户订单号获取订单
func (s *orderService) GetOrderByTradeNo(tradeNo string) (*model.Order, error) {
	order, err := s.orderRepo.GetByTradeNo(tradeNo)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	return order, nil
}

// ListUserOrders 获取用户订单列表
func (s *orderService) ListUserOrders(userID uint, page, pageSize int) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.orderRepo.ListByUserID(userID, page, pageSize)
}

// ListOrders 订单列表（管理员）
func (s *orderService) ListOrders(page, pageSize int, status int, userID uint) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.orderRepo.List(page, pageSize, status, userID)
}

// ConfirmPayment 确认支付（模拟支付回调）
func (s *orderService) ConfirmPayment(tradeNo string) error {
	order, err := s.orderRepo.GetByTradeNo(tradeNo)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return errors.New("order not found")
	}

	// 如果已支付，直接返回（幂等）
	if order.IsPaid() {
		return nil
	}

	// 如果不是待支付状态，返回错误
	if !order.IsPending() {
		return errors.New("order status is not pending")
	}

	// 使用事务确保数据一致性
	tx := database.Get().Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// 确保事务提交或回滚
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 创建带事务的仓储实例
	txOrderRepo := repository.NewOrderRepositoryWithTx(tx)
	txUserRepo := repository.NewUserRepositoryWithTx(tx)

	// 更新订单状态为已支付
	now := time.Now()
	if err := txOrderRepo.UpdateStatus(order.ID, model.OrderStatusPaid, &now); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// 为用户增加流量和时长
	userService := NewUserService(txUserRepo)
	if err := userService.AddTraffic(order.UserID, order.Plan.Traffic, order.Plan.DurationDays); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to add traffic to user: %w", err)
	}

	// 如果使用了优惠券，增加使用次数
	if order.CouponCode != "" && s.couponRepo != nil {
		coupon, err := s.couponRepo.GetByCode(order.CouponCode)
		if err == nil && coupon != nil {
			_ = s.couponRepo.IncrementUsedCount(coupon.ID)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CancelOrder 取消订单
func (s *orderService) CancelOrder(orderID uint, userID uint) error {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return errors.New("order not found")
	}

	// 检查是否是用户自己的订单
	if order.UserID != userID {
		return errors.New("order does not belong to user")
	}

	// 只有待支付的订单可以取消
	if !order.IsPending() {
		return errors.New("only pending orders can be cancelled")
	}

	return s.orderRepo.UpdateStatus(orderID, model.OrderStatusCancelled, nil)
}

// AssignOrder 管理员为用户分配订单
func (s *orderService) AssignOrder(email string, planID uint, period string, amount float64) (*model.Order, error) {
	// 根据邮箱查找用户
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 检查套餐是否存在
	plan, err := s.planRepo.GetByID(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}
	if plan == nil {
		return nil, errors.New("plan not found")
	}

	// 生成商户订单号
	tradeNo, err := generateTradeNo()
	if err != nil {
		return nil, fmt.Errorf("failed to generate trade no: %w", err)
	}

	// 如果未指定金额，使用套餐价格
	if amount <= 0 {
		amount = plan.Price
	}

	// 创建订单（管理员分配的订单直接标记为已支付）
	now := time.Now()
	order := &model.Order{
		UserID:        user.ID,
		PlanID:        planID,
		Amount:        amount,
		ActualAmount:  amount,
		Status:        model.OrderStatusPaid,
		PaymentMethod: "manual",
		TradeNo:       tradeNo,
		PaidAt:        &now,
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 为用户增加流量和时长
	userService := NewUserService(s.userRepo)
	if err := userService.AddTraffic(user.ID, plan.Traffic, plan.DurationDays); err != nil {
		return nil, fmt.Errorf("failed to add traffic to user: %w", err)
	}

	// 重新加载订单（包含套餐信息）
	return s.orderRepo.GetByID(order.ID)
}

// generateTradeNo 生成商户订单号
func generateTradeNo() (string, error) {
	// 格式：P + 时间戳 + 6位随机数
	timestamp := time.Now().Format("20060102150405")
	random, err := utils.GenerateRandomString(6)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("P%s%s", timestamp, random), nil
}
