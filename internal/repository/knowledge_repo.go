package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// KnowledgeRepository 知识库仓储接口
type KnowledgeRepository interface {
	Create(knowledge *model.Knowledge) error
	GetByID(id uint) (*model.Knowledge, error)
	Update(knowledge *model.Knowledge) error
	Delete(id uint) error
	List(page, pageSize int, category string) ([]model.Knowledge, int64, error)
	ListVisible(category, language string) ([]model.Knowledge, error)
	GetCategories() ([]string, error)
}

type knowledgeRepository struct {
	db *gorm.DB
}

// NewKnowledgeRepository 创建知识库仓储
func NewKnowledgeRepository() KnowledgeRepository {
	return &knowledgeRepository{
		db: database.Get(),
	}
}

// Create 创建知识库文章
func (r *knowledgeRepository) Create(knowledge *model.Knowledge) error {
	return r.db.Create(knowledge).Error
}

// GetByID 根据ID获取知识库文章
func (r *knowledgeRepository) GetByID(id uint) (*model.Knowledge, error) {
	var knowledge model.Knowledge
	err := r.db.First(&knowledge, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &knowledge, nil
}

// Update 更新知识库文章
func (r *knowledgeRepository) Update(knowledge *model.Knowledge) error {
	return r.db.Save(knowledge).Error
}

// Delete 删除知识库文章
func (r *knowledgeRepository) Delete(id uint) error {
	return r.db.Delete(&model.Knowledge{}, id).Error
}

// List 知识库列表（管理员）
func (r *knowledgeRepository) List(page, pageSize int, category string) ([]model.Knowledge, int64, error) {
	var knowledges []model.Knowledge
	var total int64

	query := r.db.Model(&model.Knowledge{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort ASC, id DESC").Find(&knowledges).Error; err != nil {
		return nil, 0, err
	}

	return knowledges, total, nil
}

// ListVisible 获取可见知识库列表（用户端）
func (r *knowledgeRepository) ListVisible(category, language string) ([]model.Knowledge, error) {
	var knowledges []model.Knowledge

	query := r.db.Where("show = ?", 1)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if language != "" {
		query = query.Where("language = ?", language)
	}

	err := query.Order("sort ASC, id DESC").Find(&knowledges).Error
	if err != nil {
		return nil, err
	}
	return knowledges, nil
}

// GetCategories 获取所有分类
func (r *knowledgeRepository) GetCategories() ([]string, error) {
	var categories []string
	err := r.db.Model(&model.Knowledge{}).Distinct("category").Pluck("category", &categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}
