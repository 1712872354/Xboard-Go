package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// SubscribeTemplateRepository 订阅模板仓储接口
type SubscribeTemplateRepository interface {
	Create(template *model.SubscribeTemplate) error
	GetByID(id uint) (*model.SubscribeTemplate, error)
	GetByName(name string) (*model.SubscribeTemplate, error)
	Update(template *model.SubscribeTemplate) error
	Delete(id uint) error
	List(page, pageSize int, templateType string) ([]model.SubscribeTemplate, int64, error)
	ListEnabled(templateType string) ([]model.SubscribeTemplate, error)
}

type subscribeTemplateRepository struct {
	db *gorm.DB
}

// NewSubscribeTemplateRepository 创建订阅模板仓储
func NewSubscribeTemplateRepository() SubscribeTemplateRepository {
	return &subscribeTemplateRepository{
		db: database.Get(),
	}
}

// Create 创建订阅模板
func (r *subscribeTemplateRepository) Create(template *model.SubscribeTemplate) error {
	return r.db.Create(template).Error
}

// GetByID 根据ID获取订阅模板
func (r *subscribeTemplateRepository) GetByID(id uint) (*model.SubscribeTemplate, error) {
	var template model.SubscribeTemplate
	err := r.db.First(&template, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// GetByName 根据名称获取订阅模板
func (r *subscribeTemplateRepository) GetByName(name string) (*model.SubscribeTemplate, error) {
	var template model.SubscribeTemplate
	err := r.db.Where("name = ?", name).First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// Update 更新订阅模板
func (r *subscribeTemplateRepository) Update(template *model.SubscribeTemplate) error {
	return r.db.Save(template).Error
}

// Delete 删除订阅模板
func (r *subscribeTemplateRepository) Delete(id uint) error {
	return r.db.Delete(&model.SubscribeTemplate{}, id).Error
}

// List 订阅模板列表
func (r *subscribeTemplateRepository) List(page, pageSize int, templateType string) ([]model.SubscribeTemplate, int64, error) {
	var templates []model.SubscribeTemplate
	var total int64

	query := r.db.Model(&model.SubscribeTemplate{})

	if templateType != "" {
		query = query.Where("type = ?", templateType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort ASC, id DESC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// ListEnabled 获取已启用的订阅模板列表
func (r *subscribeTemplateRepository) ListEnabled(templateType string) ([]model.SubscribeTemplate, error) {
	var templates []model.SubscribeTemplate

	query := r.db.Where("status = ?", 1)

	if templateType != "" {
		query = query.Where("type = ?", templateType)
	}

	err := query.Order("sort ASC, id DESC").Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}
