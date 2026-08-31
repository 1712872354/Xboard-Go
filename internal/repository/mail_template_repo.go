package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// MailTemplateRepository 邮件模板仓储接口
type MailTemplateRepository interface {
	Create(template *model.MailTemplate) error
	GetByID(id uint) (*model.MailTemplate, error)
	GetByName(name string) (*model.MailTemplate, error)
	Update(template *model.MailTemplate) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.MailTemplate, int64, error)
}

type mailTemplateRepository struct {
	db *gorm.DB
}

// NewMailTemplateRepository 创建邮件模板仓储
func NewMailTemplateRepository() MailTemplateRepository {
	return &mailTemplateRepository{
		db: database.Get(),
	}
}

// Create 创建邮件模板
func (r *mailTemplateRepository) Create(template *model.MailTemplate) error {
	return r.db.Create(template).Error
}

// GetByID 根据ID获取邮件模板
func (r *mailTemplateRepository) GetByID(id uint) (*model.MailTemplate, error) {
	var template model.MailTemplate
	err := r.db.First(&template, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// GetByName 根据名称获取邮件模板
func (r *mailTemplateRepository) GetByName(name string) (*model.MailTemplate, error) {
	var template model.MailTemplate
	err := r.db.Where("name = ?", name).First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// Update 更新邮件模板
func (r *mailTemplateRepository) Update(template *model.MailTemplate) error {
	return r.db.Save(template).Error
}

// Delete 删除邮件模板
func (r *mailTemplateRepository) Delete(id uint) error {
	return r.db.Delete(&model.MailTemplate{}, id).Error
}

// List 邮件模板列表
func (r *mailTemplateRepository) List(page, pageSize int) ([]model.MailTemplate, int64, error) {
	var templates []model.MailTemplate
	var total int64

	query := r.db.Model(&model.MailTemplate{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}
