package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// PlanRepository 套餐仓储接口
type PlanRepository interface {
	Create(plan *model.Plan) error
	GetByID(id uint) (*model.Plan, error)
	Update(plan *model.Plan) error
	Delete(id uint) error
	List(page, pageSize int, includeDisabled bool) ([]model.Plan, int64, error)
	ListActive() ([]model.Plan, error)
}

type planRepository struct {
	db *gorm.DB
}

// NewPlanRepository 创建套餐仓储
func NewPlanRepository() PlanRepository {
	return &planRepository{
		db: database.Get(),
	}
}

// Create 创建套餐
func (r *planRepository) Create(plan *model.Plan) error {
	return r.db.Create(plan).Error
}

// GetByID 根据ID获取套餐
func (r *planRepository) GetByID(id uint) (*model.Plan, error) {
	var plan model.Plan
	err := r.db.First(&plan, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}

// Update 更新套餐
func (r *planRepository) Update(plan *model.Plan) error {
	return r.db.Save(plan).Error
}

// Delete 删除套餐
func (r *planRepository) Delete(id uint) error {
	return r.db.Delete(&model.Plan{}, id).Error
}

// List 套餐列表（管理员）
func (r *planRepository) List(page, pageSize int, includeDisabled bool) ([]model.Plan, int64, error) {
	var plans []model.Plan
	var total int64

	query := r.db.Model(&model.Plan{})

	if !includeDisabled {
		query = query.Where("status = ?", 1)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id ASC").Find(&plans).Error; err != nil {
		return nil, 0, err
	}

	return plans, total, nil
}

// ListActive 获取上架的套餐列表
func (r *planRepository) ListActive() ([]model.Plan, error) {
	var plans []model.Plan
	err := r.db.Where("status = ?", 1).Order("id ASC").Find(&plans).Error
	return plans, err
}
