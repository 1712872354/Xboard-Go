package repository

import (
	"errors"
	"time"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// RedeemCodeRepository 兑换码仓储接口
type RedeemCodeRepository interface {
	Create(code *model.RedeemCode) error
	BatchCreate(codes []model.RedeemCode) error
	GetByID(id uint) (*model.RedeemCode, error)
	GetByCode(code string) (*model.RedeemCode, error)
	Update(code *model.RedeemCode) error
	Delete(id uint) error
	List(page, pageSize int, status int, planID uint) ([]model.RedeemCode, int64, error)
	// Redeem 使用兑换码（事务：标记已使用）
	Redeem(code string, userID uint) (*model.RedeemCode, error)
	CountByStatus(status int) (int64, error)
}

type redeemCodeRepository struct {
	db *gorm.DB
}

// NewRedeemCodeRepository 创建兑换码仓储
func NewRedeemCodeRepository() RedeemCodeRepository {
	return &redeemCodeRepository{
		db: database.Get(),
	}
}

// Create 创建兑换码
func (r *redeemCodeRepository) Create(code *model.RedeemCode) error {
	return r.db.Create(code).Error
}

// BatchCreate 批量创建兑换码
func (r *redeemCodeRepository) BatchCreate(codes []model.RedeemCode) error {
	if len(codes) == 0 {
		return nil
	}
	return r.db.Create(&codes).Error
}

// GetByID 根据ID获取兑换码
func (r *redeemCodeRepository) GetByID(id uint) (*model.RedeemCode, error) {
	var code model.RedeemCode
	err := r.db.Preload("Plan").First(&code, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &code, nil
}

// GetByCode 根据兑换码获取
func (r *redeemCodeRepository) GetByCode(code string) (*model.RedeemCode, error) {
	var redeemCode model.RedeemCode
	err := r.db.Where("code = ?", code).Preload("Plan").First(&redeemCode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &redeemCode, nil
}

// Update 更新兑换码
func (r *redeemCodeRepository) Update(code *model.RedeemCode) error {
	return r.db.Save(code).Error
}

// Delete 删除兑换码
func (r *redeemCodeRepository) Delete(id uint) error {
	return r.db.Delete(&model.RedeemCode{}, id).Error
}

// List 兑换码列表
func (r *redeemCodeRepository) List(page, pageSize int, status int, planID uint) ([]model.RedeemCode, int64, error) {
	var codes []model.RedeemCode
	var total int64

	query := r.db.Model(&model.RedeemCode{}).Preload("Plan")

	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if planID > 0 {
		query = query.Where("plan_id = ?", planID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&codes).Error; err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}

// Redeem 使用兑换码（事务操作，保证原子性）
func (r *redeemCodeRepository) Redeem(code string, userID uint) (*model.RedeemCode, error) {
	var redeemCode model.RedeemCode

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 查询兑换码（加行锁）
		if err := tx.Where("code = ?", code).First(&redeemCode).Error; err != nil {
			return err
		}

		// 2. 检查是否已使用
		if redeemCode.IsUsed() {
			return errors.New("redeem code already used")
		}

		// 3. 标记为已使用
		now := time.Now()
		redeemCode.Status = 1
		redeemCode.UsedBy = &userID
		redeemCode.UsedAt = &now

		result := tx.Model(&redeemCode).
			Where("id = ? AND status = ?", redeemCode.ID, 0).
			Updates(map[string]interface{}{
				"status":   1,
				"used_by":  userID,
				"used_at":  now,
			})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("redeem code already used (concurrent)")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &redeemCode, nil
}

// CountByStatus 按状态统计数量
func (r *redeemCodeRepository) CountByStatus(status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.RedeemCode{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
