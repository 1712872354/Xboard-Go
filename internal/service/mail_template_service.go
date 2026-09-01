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
	ResetMailTemplate(id uint) (*model.MailTemplate, error)
	TestMailTemplate(id uint, email string) error
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

// defaultTemplates 内置默认模板
var defaultTemplates = map[string]struct {
	Subject string
	Body    string
}{
	model.MailTemplateRegister: {
		Subject: "注册验证码 - {{site.name}}",
		Body:    `<h2>{{site.name}} - 注册验证码</h2><p>您的验证码是：<strong>{{code}}</strong></p><p>验证码 10 分钟内有效，请勿泄露给他人。</p>`,
	},
	model.MailTemplateResetPassword: {
		Subject: "重置密码 - {{site.name}}",
		Body:    `<h2>{{site.name}} - 重置密码</h2><p>您的验证码是：<strong>{{code}}</strong></p><p>验证码 10 分钟内有效，请勿泄露给他人。</p>`,
	},
	model.MailTemplateOrderCreated: {
		Subject: "订单创建成功 - {{site.name}}",
		Body:    `<h2>{{site.name}} - 订单创建成功</h2><p>订单号：{{order.trade_no}}</p><p>金额：{{order.amount}}</p><p>请及时完成支付。</p>`,
	},
	model.MailTemplateOrderPaid: {
		Subject: "订单支付成功 - {{site.name}}",
		Body:    `<h2>{{site.name}} - 订单支付成功</h2><p>订单号：{{order.trade_no}}</p><p>感谢您的购买！</p>`,
	},
	model.MailTemplateTrafficWarning: {
		Subject: "流量使用提醒 - {{site.name}}",
		Body:    `<h2>{{site.name}} - 流量使用提醒</h2><p>您的流量使用已超过 80%，请注意合理使用。</p>`,
	},
	model.MailTemplateAccountExpired: {
		Subject: "服务到期提醒 - {{site.name}}",
		Body:    `<h2>{{site.name}} - 服务到期提醒</h2><p>您的服务即将到期，请及时续费。</p>`,
	},
	model.MailTemplateWelcome: {
		Subject: "欢迎加入 {{site.name}}",
		Body:    `<h2>欢迎加入 {{site.name}}</h2><p>您的账号已注册成功，祝您使用愉快！</p>`,
	},
	model.MailTemplateTicketReply: {
		Subject: "工单回复 - {{site.name}}",
		Body:    `<h2>{{site.name}} - 工单回复</h2><p>您的工单有新的回复，请前往查看。</p>`,
	},
}

// ResetMailTemplate 将模板恢复为默认值
func (s *mailTemplateService) ResetMailTemplate(id uint) (*model.MailTemplate, error) {
	template, err := s.mailTemplateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("template not found")
	}

	// 查找默认模板
	defaultTpl, ok := defaultTemplates[template.Name]
	if !ok {
		return nil, errors.New("no default template found for this template name")
	}

	template.Subject = defaultTpl.Subject
	template.Body = defaultTpl.Body

	if err := s.mailTemplateRepo.Update(template); err != nil {
		return nil, err
	}

	return template, nil
}

// TestMailTemplate 发送测试邮件
func (s *mailTemplateService) TestMailTemplate(id uint, email string) error {
	template, err := s.mailTemplateRepo.GetByID(id)
	if err != nil {
		return err
	}
	if template == nil {
		return errors.New("template not found")
	}

	mailSvc := NewMailService()
	return mailSvc.SendEmail(email, template.Subject, template.Name, map[string]interface{}{
		"site.name": "XBoard",
		"site.url":  "",
		"code":      "888888",
	})
}
