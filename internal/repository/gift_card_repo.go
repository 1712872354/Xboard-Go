package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// GiftCardTemplateRepository 礼品卡模板仓储接口
type GiftCardTemplateRepository interface {
	Create(template *model.GiftCardTemplate) error
	GetByID(id uint) (*model.GiftCardTemplate, error)
	Update(template *model.GiftCardTemplate) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.GiftCardTemplate, int64, error)
}

type giftCardTemplateRepository struct {
	db *gorm.DB
}

// NewGiftCardTemplateRepository 创建礼品卡模板仓储
func NewGiftCardTemplateRepository() GiftCardTemplateRepository {
	return &giftCardTemplateRepository{
		db: database.Get(),
	}
}

// Create 创建礼品卡模板
func (r *giftCardTemplateRepository) Create(template *model.GiftCardTemplate) error {
	return r.db.Create(template).Error
}

// GetByID 根据ID获取礼品卡模板
func (r *giftCardTemplateRepository) GetByID(id uint) (*model.GiftCardTemplate, error) {
	var template model.GiftCardTemplate
	err := r.db.First(&template, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// Update 更新礼品卡模板
func (r *giftCardTemplateRepository) Update(template *model.GiftCardTemplate) error {
	return r.db.Save(template).Error
}

// Delete 删除礼品卡模板
func (r *giftCardTemplateRepository) Delete(id uint) error {
	return r.db.Delete(&model.GiftCardTemplate{}, id).Error
}

// List 礼品卡模板列表
func (r *giftCardTemplateRepository) List(page, pageSize int) ([]model.GiftCardTemplate, int64, error) {
	var templates []model.GiftCardTemplate
	var total int64

	query := r.db.Model(&model.GiftCardTemplate{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// GiftCardCodeRepository 礼品卡码仓储接口
type GiftCardCodeRepository interface {
	Create(code *model.GiftCardCode) error
	CreateBatch(codes []model.GiftCardCode) error
	GetByID(id uint) (*model.GiftCardCode, error)
	GetByCode(code string) (*model.GiftCardCode, error)
	Update(code *model.GiftCardCode) error
	Delete(id uint) error
	List(page, pageSize int, templateID uint, status int) ([]model.GiftCardCode, int64, error)
	CountByTemplate(templateID uint) (int64, error)
	CountUsedByTemplate(templateID uint) (int64, error)
}

type giftCardCodeRepository struct {
	db *gorm.DB
}

// NewGiftCardCodeRepository 创建礼品卡码仓储
func NewGiftCardCodeRepository() GiftCardCodeRepository {
	return &giftCardCodeRepository{
		db: database.Get(),
	}
}

// Create 创建礼品卡码
func (r *giftCardCodeRepository) Create(code *model.GiftCardCode) error {
	return r.db.Create(code).Error
}

// CreateBatch 批量创建礼品卡码
func (r *giftCardCodeRepository) CreateBatch(codes []model.GiftCardCode) error {
	return r.db.CreateInBatches(codes, 100).Error
}

// GetByID 根据ID获取礼品卡码
func (r *giftCardCodeRepository) GetByID(id uint) (*model.GiftCardCode, error) {
	var code model.GiftCardCode
	err := r.db.First(&code, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &code, nil
}

// GetByCode 根据礼品码获取礼品卡
func (r *giftCardCodeRepository) GetByCode(code string) (*model.GiftCardCode, error) {
	var giftCard model.GiftCardCode
	err := r.db.Where("code = ?", code).First(&giftCard).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &giftCard, nil
}

// Update 更新礼品卡码
func (r *giftCardCodeRepository) Update(code *model.GiftCardCode) error {
	return r.db.Save(code).Error
}

// Delete 删除礼品卡码
func (r *giftCardCodeRepository) Delete(id uint) error {
	return r.db.Delete(&model.GiftCardCode{}, id).Error
}

// List 礼品卡码列表
func (r *giftCardCodeRepository) List(page, pageSize int, templateID uint, status int) ([]model.GiftCardCode, int64, error) {
	var codes []model.GiftCardCode
	var total int64

	query := r.db.Model(&model.GiftCardCode{})

	if templateID > 0 {
		query = query.Where("template_id = ?", templateID)
	}

	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&codes).Error; err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}

// CountByTemplate 统计模板下的礼品码数量
func (r *giftCardCodeRepository) CountByTemplate(templateID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.GiftCardCode{}).Where("template_id = ?", templateID).Count(&count).Error
	return count, err
}

// CountUsedByTemplate 统计模板下已使用的礼品码数量
func (r *giftCardCodeRepository) CountUsedByTemplate(templateID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.GiftCardCode{}).Where("template_id = ? AND status = 1", templateID).Count(&count).Error
	return count, err
}

// GiftCardUsageRepository 礼品卡使用记录仓储接口
type GiftCardUsageRepository interface {
	Create(usage *model.GiftCardUsage) error
	List(page, pageSize int, userID uint) ([]model.GiftCardUsage, int64, error)
}

type giftCardUsageRepository struct {
	db *gorm.DB
}

// NewGiftCardUsageRepository 创建礼品卡使用记录仓储
func NewGiftCardUsageRepository() GiftCardUsageRepository {
	return &giftCardUsageRepository{
		db: database.Get(),
	}
}

// Create 创建使用记录
func (r *giftCardUsageRepository) Create(usage *model.GiftCardUsage) error {
	return r.db.Create(usage).Error
}

// List 使用记录列表
func (r *giftCardUsageRepository) List(page, pageSize int, userID uint) ([]model.GiftCardUsage, int64, error) {
	var usages []model.GiftCardUsage
	var total int64

	query := r.db.Model(&model.GiftCardUsage{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&usages).Error; err != nil {
		return nil, 0, err
	}

	return usages, total, nil
}
