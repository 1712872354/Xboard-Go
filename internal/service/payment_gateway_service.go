package service

import (
	"fmt"
	"time"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// PaymentGatewayService 支付网关服务接口
type PaymentGatewayService interface {
	// CreateGateway 创建支付网关
	CreateGateway(name, icon, payment string, config map[string]interface{}, notifyDomain string) (*model.Payment, error)
	// GetGateway 获取支付网关
	GetGateway(id uint) (*model.Payment, error)
	// UpdateGateway 更新支付网关
	UpdateGateway(id uint, name, icon, payment string, config map[string]interface{}, notifyDomain string) (*model.Payment, error)
	// DeleteGateway 删除支付网关
	DeleteGateway(id uint) error
	// ListGateways 获取支付网关列表
	ListGateways() ([]model.Payment, error)
	// UpdateGatewayStatus 更新支付网关状态
	UpdateGatewayStatus(id uint, status int) error
	// UpdateGatewaySort 更新支付网关排序
	UpdateGatewaySort(id uint, sort int) error
	// GetGatewayConfig 获取支付网关配置
	GetGatewayConfig(id uint) (map[string]interface{}, error)
	// GetGatewayByPayment 根据支付方式获取网关
	GetGatewayByPayment(payment string) (*model.Payment, error)
}

// paymentGatewayService 支付网关服务实现
type paymentGatewayService struct {
	paymentRepo repository.PaymentRepository
}

// NewPaymentGatewayService 创建支付网关服务
func NewPaymentGatewayService(paymentRepo repository.PaymentRepository) PaymentGatewayService {
	return &paymentGatewayService{
		paymentRepo: paymentRepo,
	}
}

// CreateGateway 创建支付网关
func (s *paymentGatewayService) CreateGateway(name, icon, payment string, config map[string]interface{}, notifyDomain string) (*model.Payment, error) {
	// 检查支付方式是否已存在
	existing, err := s.paymentRepo.GetByPayment(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing payment: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("payment method %s already exists", payment)
	}

	// 创建支付网关
	gateway := &model.Payment{
		Name:          name,
		Icon:          icon,
		Payment:       payment,
		Config:        config,
		NotifyDomain:  notifyDomain,
		Status:        1,
		Sort:          0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.paymentRepo.Create(gateway); err != nil {
		return nil, fmt.Errorf("failed to create payment gateway: %w", err)
	}

	return gateway, nil
}

// GetGateway 获取支付网关
func (s *paymentGatewayService) GetGateway(id uint) (*model.Payment, error) {
	gateway, err := s.paymentRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if gateway == nil {
		return nil, fmt.Errorf("payment gateway not found")
	}
	return gateway, nil
}

// UpdateGateway 更新支付网关
func (s *paymentGatewayService) UpdateGateway(id uint, name, icon, payment string, config map[string]interface{}, notifyDomain string) (*model.Payment, error) {
	gateway, err := s.paymentRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if gateway == nil {
		return nil, fmt.Errorf("payment gateway not found")
	}

	// 检查支付方式是否被其他网关使用
	if payment != gateway.Payment {
		existing, err := s.paymentRepo.GetByPayment(payment)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing payment: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("payment method %s already exists", payment)
		}
	}

	// 更新字段
	gateway.Name = name
	gateway.Icon = icon
	gateway.Payment = payment
	gateway.Config = config
	gateway.NotifyDomain = notifyDomain
	gateway.UpdatedAt = time.Now()

	if err := s.paymentRepo.Update(gateway); err != nil {
		return nil, fmt.Errorf("failed to update payment gateway: %w", err)
	}

	return gateway, nil
}

// DeleteGateway 删除支付网关
func (s *paymentGatewayService) DeleteGateway(id uint) error {
	gateway, err := s.paymentRepo.GetByID(id)
	if err != nil {
		return err
	}
	if gateway == nil {
		return fmt.Errorf("payment gateway not found")
	}

	return s.paymentRepo.Delete(id)
}

// ListGateways 获取支付网关列表
func (s *paymentGatewayService) ListGateways() ([]model.Payment, error) {
	return s.paymentRepo.List()
}

// UpdateGatewayStatus 更新支付网关状态
func (s *paymentGatewayService) UpdateGatewayStatus(id uint, status int) error {
	gateway, err := s.paymentRepo.GetByID(id)
	if err != nil {
		return err
	}
	if gateway == nil {
		return fmt.Errorf("payment gateway not found")
	}

	gateway.Status = status
	gateway.UpdatedAt = time.Now()

	return s.paymentRepo.Update(gateway)
}

// UpdateGatewaySort 更新支付网关排序
func (s *paymentGatewayService) UpdateGatewaySort(id uint, sort int) error {
	gateway, err := s.paymentRepo.GetByID(id)
	if err != nil {
		return err
	}
	if gateway == nil {
		return fmt.Errorf("payment gateway not found")
	}

	gateway.Sort = sort
	gateway.UpdatedAt = time.Now()

	return s.paymentRepo.Update(gateway)
}

// GetGatewayConfig 获取支付网关配置
func (s *paymentGatewayService) GetGatewayConfig(id uint) (map[string]interface{}, error) {
	gateway, err := s.paymentRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if gateway == nil {
		return nil, fmt.Errorf("payment gateway not found")
	}

	return gateway.Config, nil
}

// GetGatewayByPayment 根据支付方式获取网关
func (s *paymentGatewayService) GetGatewayByPayment(payment string) (*model.Payment, error) {
	return s.paymentRepo.GetByPayment(payment)
}
