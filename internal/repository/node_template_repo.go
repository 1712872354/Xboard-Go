package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// NodeTemplateRepository 节点模板仓储接口
type NodeTemplateRepository interface {
	Create(template *model.NodeTemplate) error
	GetByID(id uint) (*model.NodeTemplate, error)
	Update(template *model.NodeTemplate) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.NodeTemplate, int64, error)
}

type nodeTemplateRepository struct {
	db *gorm.DB
}

// NewNodeTemplateRepository 创建节点模板仓储
func NewNodeTemplateRepository() NodeTemplateRepository {
	return &nodeTemplateRepository{
		db: database.Get(),
	}
}

// Create 创建节点模板
func (r *nodeTemplateRepository) Create(template *model.NodeTemplate) error {
	return r.db.Create(template).Error
}

// GetByID 根据ID获取节点模板
func (r *nodeTemplateRepository) GetByID(id uint) (*model.NodeTemplate, error) {
	var template model.NodeTemplate
	err := r.db.First(&template, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// Update 更新节点模板
func (r *nodeTemplateRepository) Update(template *model.NodeTemplate) error {
	return r.db.Save(template).Error
}

// Delete 删除节点模板
func (r *nodeTemplateRepository) Delete(id uint) error {
	return r.db.Delete(&model.NodeTemplate{}, id).Error
}

// List 节点模板列表
func (r *nodeTemplateRepository) List(page, pageSize int) ([]model.NodeTemplate, int64, error) {
	var templates []model.NodeTemplate
	var total int64

	query := r.db.Model(&model.NodeTemplate{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}
