package service

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// NodeTemplateService 节点模板服务接口
type NodeTemplateService interface {
	ListNodeTemplates(page, pageSize int) ([]model.NodeTemplate, int64, error)
	CreateNodeTemplate(name, nodeType, serverInfo, description string) (*model.NodeTemplate, error)
	GetNodeTemplateByID(id uint) (*model.NodeTemplate, error)
	UpdateNodeTemplate(id uint, name, nodeType, serverInfo, description string) (*model.NodeTemplate, error)
	DeleteNodeTemplate(id uint) error
}

type nodeTemplateService struct {
	templateRepo repository.NodeTemplateRepository
}

// NewNodeTemplateService 创建节点模板服务
func NewNodeTemplateService(templateRepo repository.NodeTemplateRepository) NodeTemplateService {
	return &nodeTemplateService{
		templateRepo: templateRepo,
	}
}

// ListNodeTemplates 节点模板列表
func (s *nodeTemplateService) ListNodeTemplates(page, pageSize int) ([]model.NodeTemplate, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.templateRepo.List(page, pageSize)
}

// CreateNodeTemplate 创建节点模板
func (s *nodeTemplateService) CreateNodeTemplate(name, nodeType, serverInfo, description string) (*model.NodeTemplate, error) {
	if name == "" {
		return nil, errors.New("template name is required")
	}
	if nodeType == "" {
		return nil, errors.New("template type is required")
	}

	template := &model.NodeTemplate{
		Name:        name,
		Type:        nodeType,
		ServerInfo:  serverInfo,
		Description: description,
	}

	if err := s.templateRepo.Create(template); err != nil {
		return nil, err
	}

	return template, nil
}

// GetNodeTemplateByID 根据ID获取节点模板
func (s *nodeTemplateService) GetNodeTemplateByID(id uint) (*model.NodeTemplate, error) {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("node template not found")
	}
	return template, nil
}

// UpdateNodeTemplate 更新节点模板
func (s *nodeTemplateService) UpdateNodeTemplate(id uint, name, nodeType, serverInfo, description string) (*model.NodeTemplate, error) {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("node template not found")
	}

	if name != "" {
		template.Name = name
	}
	if nodeType != "" {
		template.Type = nodeType
	}
	if serverInfo != "" {
		template.ServerInfo = serverInfo
	}
	if description != "" {
		template.Description = description
	}

	if err := s.templateRepo.Update(template); err != nil {
		return nil, err
	}

	return template, nil
}

// DeleteNodeTemplate 删除节点模板
func (s *nodeTemplateService) DeleteNodeTemplate(id uint) error {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return err
	}
	if template == nil {
		return errors.New("node template not found")
	}

	return s.templateRepo.Delete(id)
}
