package repository

import (
	"errors"
	"fmt"
	"time"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// OrderRepository 订单仓储接口
type OrderRepository interface {
	Create(order *model.Order) error
	GetByID(id uint) (*model.Order, error)
	GetByTradeNo(tradeNo string) (*model.Order, error)
	Update(order *model.Order) error
	ListByUserID(userID uint, page, pageSize int) ([]model.Order, int64, error)
	List(page, pageSize int, status int, userID uint) ([]model.Order, int64, error)
	UpdateStatus(orderID uint, status int, paidAt *time.Time) error
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓储
func NewOrderRepository() OrderRepository {
	return &orderRepository{
		db: database.Get(),
	}
}

// NewOrderRepositoryWithTx 创建带事务的订单仓储
func NewOrderRepositoryWithTx(tx *gorm.DB) OrderRepository {
	return &orderRepository{
		db: tx,
	}
}

// Create 创建订单
func (r *orderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

// GetByID 根据ID获取订单
func (r *orderRepository) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Plan").First(&order, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetByTradeNo 根据商户订单号获取订单
func (r *orderRepository) GetByTradeNo(tradeNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.Where("trade_no = ?", tradeNo).Preload("Plan").First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// Update 更新订单
func (r *orderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

// ListByUserID 获取用户订单列表
func (r *orderRepository) ListByUserID(userID uint, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Plan").Offset(offset).Limit(pageSize).
		Order("id DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// List 订单列表（管理员）
func (r *orderRepository) List(page, pageSize int, status int, userID uint) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{})

	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Plan").Preload("User").Offset(offset).Limit(pageSize).
		Order("id DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// UpdateStatus 更新订单状态（幂等操作）
func (r *orderRepository) UpdateStatus(orderID uint, status int, paidAt *time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var order model.Order
		if err := tx.First(&order, orderID).Error; err != nil {
			return err
		}

		// 如果已经是目标状态，直接返回（幂等）
		if order.Status == status {
			return nil
		}

		updates := map[string]interface{}{
			"status": status,
		}
		if paidAt != nil {
			updates["paid_at"] = paidAt
		}

		result := tx.Model(&order).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("no rows affected")
		}

		return nil
	})
}
