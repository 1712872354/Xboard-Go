package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// PaymentRepository 支付方式仓储接口
type PaymentRepository interface {
	Create(payment *model.Payment) error
	GetByID(id uint) (*model.Payment, error)
	GetByPayment(payment string) (*model.Payment, error)
	Update(payment *model.Payment) error
	Delete(id uint) error
	List() ([]model.Payment, error)
}

type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository 创建支付方式仓储
func NewPaymentRepository() PaymentRepository {
	return &paymentRepository{
		db: database.Get(),
	}
}

// Create 创建支付方式
func (r *paymentRepository) Create(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

// GetByID 根据ID获取支付方式
func (r *paymentRepository) GetByID(id uint) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.First(&payment, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// GetByPayment 根据支付方式标识获取
func (r *paymentRepository) GetByPayment(payment string) (*model.Payment, error) {
	var p model.Payment
	err := r.db.Where("payment = ?", payment).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// Update 更新支付方式
func (r *paymentRepository) Update(payment *model.Payment) error {
	return r.db.Save(payment).Error
}

// Delete 删除支付方式
func (r *paymentRepository) Delete(id uint) error {
	return r.db.Delete(&model.Payment{}, id).Error
}

// List 获取支付方式列表
func (r *paymentRepository) List() ([]model.Payment, error) {
	var payments []model.Payment
	err := r.db.Order("sort ASC, id ASC").Find(&payments).Error
	return payments, err
}
