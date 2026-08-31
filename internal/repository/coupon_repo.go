package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// CouponRepository 优惠券仓储接口
type CouponRepository interface {
	Create(coupon *model.Coupon) error
	GetByID(id uint) (*model.Coupon, error)
	GetByCode(code string) (*model.Coupon, error)
	Update(coupon *model.Coupon) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.Coupon, int64, error)
	IncrementUsedCount(id uint) error
}

type couponRepository struct {
	db *gorm.DB
}

// NewCouponRepository 创建优惠券仓储
func NewCouponRepository() CouponRepository {
	return &couponRepository{
		db: database.Get(),
	}
}

// Create 创建优惠券
func (r *couponRepository) Create(coupon *model.Coupon) error {
	return r.db.Create(coupon).Error
}

// GetByID 根据ID获取优惠券
func (r *couponRepository) GetByID(id uint) (*model.Coupon, error) {
	var coupon model.Coupon
	err := r.db.First(&coupon, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &coupon, nil
}

// GetByCode 根据优惠码获取优惠券
func (r *couponRepository) GetByCode(code string) (*model.Coupon, error) {
	var coupon model.Coupon
	err := r.db.Where("code = ?", code).First(&coupon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &coupon, nil
}

// Update 更新优惠券
func (r *couponRepository) Update(coupon *model.Coupon) error {
	return r.db.Save(coupon).Error
}

// Delete 删除优惠券
func (r *couponRepository) Delete(id uint) error {
	return r.db.Delete(&model.Coupon{}, id).Error
}

// List 优惠券列表
func (r *couponRepository) List(page, pageSize int) ([]model.Coupon, int64, error) {
	var coupons []model.Coupon
	var total int64

	query := r.db.Model(&model.Coupon{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&coupons).Error; err != nil {
		return nil, 0, err
	}

	return coupons, total, nil
}

// IncrementUsedCount 增加使用次数
func (r *couponRepository) IncrementUsedCount(id uint) error {
	return r.db.Model(&model.Coupon{}).Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}
