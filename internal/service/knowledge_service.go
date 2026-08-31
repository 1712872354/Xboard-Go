package service

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// KnowledgeService 知识库服务接口
type KnowledgeService interface {
	Create(category, title, content, language string, show, sort int) (*model.Knowledge, error)
	GetByID(id uint) (*model.Knowledge, error)
	Update(id uint, category, title, content, language string, show, sort int) (*model.Knowledge, error)
	Delete(id uint) error
	List(page, pageSize int, category string) ([]model.Knowledge, int64, error)
	ListVisible(category, language string) ([]model.Knowledge, error)
	GetCategories() ([]string, error)
}

type knowledgeService struct {
	knowledgeRepo repository.KnowledgeRepository
}

// NewKnowledgeService 创建知识库服务
func NewKnowledgeService(knowledgeRepo repository.KnowledgeRepository) KnowledgeService {
	return &knowledgeService{
		knowledgeRepo: knowledgeRepo,
	}
}

// Create 创建知识库文章
func (s *knowledgeService) Create(category, title, content, language string, show, sort int) (*model.Knowledge, error) {
	knowledge := &model.Knowledge{
		Category: category,
		Title:    title,
		Content:  content,
		Language: language,
		Show:     show,
		Sort:     sort,
	}

	if err := s.knowledgeRepo.Create(knowledge); err != nil {
		return nil, err
	}

	return knowledge, nil
}

// GetByID 根据ID获取知识库文章
func (s *knowledgeService) GetByID(id uint) (*model.Knowledge, error) {
	knowledge, err := s.knowledgeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if knowledge == nil {
		return nil, errors.New("knowledge not found")
	}
	return knowledge, nil
}

// Update 更新知识库文章
func (s *knowledgeService) Update(id uint, category, title, content, language string, show, sort int) (*model.Knowledge, error) {
	knowledge, err := s.knowledgeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if knowledge == nil {
		return nil, errors.New("knowledge not found")
	}

	knowledge.Category = category
	knowledge.Title = title
	knowledge.Content = content
	knowledge.Language = language
	knowledge.Show = show
	knowledge.Sort = sort

	if err := s.knowledgeRepo.Update(knowledge); err != nil {
		return nil, err
	}

	return knowledge, nil
}

// Delete 删除知识库文章
func (s *knowledgeService) Delete(id uint) error {
	knowledge, err := s.knowledgeRepo.GetByID(id)
	if err != nil {
		return err
	}
	if knowledge == nil {
		return errors.New("knowledge not found")
	}

	return s.knowledgeRepo.Delete(id)
}

// List 知识库列表（管理员）
func (s *knowledgeService) List(page, pageSize int, category string) ([]model.Knowledge, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.knowledgeRepo.List(page, pageSize, category)
}

// ListVisible 获取可见知识库列表（用户端）
func (s *knowledgeService) ListVisible(category, language string) ([]model.Knowledge, error) {
	return s.knowledgeRepo.ListVisible(category, language)
}

// GetCategories 获取所有分类
func (s *knowledgeService) GetCategories() ([]string, error) {
	return s.knowledgeRepo.GetCategories()
}
