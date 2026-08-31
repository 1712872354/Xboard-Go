package service

import (
	"errors"
	"strings"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// MailTemplateService 邮件模板服务接口
type MailTemplateService interface {
	Create(name, subject, body, language, remark string) (*model.MailTemplate, error)
	GetByID(id uint) (*model.MailTemplate, error)
	GetByName(name string) (*model.MailTemplate, error)
	Update(id uint, name, subject, body, language, remark string) (*model.MailTemplate, error)
	Delete(id uint) error
	List(page, pageSize int) ([]model.MailTemplate, int64, error)
	RenderTemplate(name string, data map[string]string) (subject string, body string, err error)
}

type mailTemplateService struct {
	mailTemplateRepo repository.MailTemplateRepository
}

// NewMailTemplateService 创建邮件模板服务
func NewMailTemplateService(mailTemplateRepo repository.MailTemplateRepository) MailTemplateService {
	return &mailTemplateService{
		mailTemplateRepo: mailTemplateRepo,
	}
}

// Create 创建邮件模板
func (s *mailTemplateService) Create(name, subject, body, language, remark string) (*model.MailTemplate, error) {
	// 检查名称是否已存在
	existing, err := s.mailTemplateRepo.GetByName(name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("template name already exists")
	}

	template := &model.MailTemplate{
		Name:     name,
		Subject:  subject,
		Body:     body,
		Language: language,
		Remark:   remark,
	}

	if err := s.mailTemplateRepo.Create(template); err != nil {
		return nil, err
	}

	return template, nil
}

// GetByID 根据ID获取邮件模板
func (s *mailTemplateService) GetByID(id uint) (*model.MailTemplate, error) {
	template, err := s.mailTemplateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("template not found")
	}
	return template, nil
}

// GetByName 根据名称获取邮件模板
func (s *mailTemplateService) GetByName(name string) (*model.MailTemplate, error) {
	template, err := s.mailTemplateRepo.GetByName(name)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("template not found")
	}
	return template, nil
}

// Update 更新邮件模板
func (s *mailTemplateService) Update(id uint, name, subject, body, language, remark string) (*model.MailTemplate, error) {
	template, err := s.mailTemplateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("template not found")
	}

	// 如果修改了名称，检查新名称是否已存在
	if name != template.Name {
		existing, err := s.mailTemplateRepo.GetByName(name)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, errors.New("template name already exists")
		}
	}

	template.Name = name
	template.Subject = subject
	template.Body = body
	template.Language = language
	template.Remark = remark

	if err := s.mailTemplateRepo.Update(template); err != nil {
		return nil, err
	}

	return template, nil
}

// Delete 删除邮件模板
func (s *mailTemplateService) Delete(id uint) error {
	template, err := s.mailTemplateRepo.GetByID(id)
	if err != nil {
		return err
	}
	if template == nil {
		return errors.New("template not found")
	}

	return s.mailTemplateRepo.Delete(id)
}

// List 邮件模板列表
func (s *mailTemplateService) List(page, pageSize int) ([]model.MailTemplate, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.mailTemplateRepo.List(page, pageSize)
}

// RenderTemplate 渲染邮件模板
func (s *mailTemplateService) RenderTemplate(name string, data map[string]string) (subject string, body string, err error) {
	template, err := s.mailTemplateRepo.GetByName(name)
	if err != nil {
		return "", "", err
	}
	if template == nil {
		return "", "", errors.New("template not found")
	}

	// 替换模板中的变量
	subject = template.Subject
	body = template.Body

	for key, value := range data {
		placeholder := "{{" + key + "}}"
		subject = strings.ReplaceAll(subject, placeholder, value)
		body = strings.ReplaceAll(body, placeholder, value)
	}

	return subject, body, nil
}
