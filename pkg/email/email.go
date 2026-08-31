package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailDriver 邮件驱动接口
type EmailDriver interface {
	Send(to, subject, body string) error
}

// SMTPConfig SMTP 配置
type SMTPConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
	FromName    string
	Encryption  string // tls, ssl, none
}

// SMTPDriver SMTP 邮件驱动
type SMTPDriver struct {
	config *SMTPConfig
}

// NewSMTPDriver 创建 SMTP 驱动
func NewSMTPDriver(config *SMTPConfig) *SMTPDriver {
	return &SMTPDriver{
		config: config,
	}
}

// Send 发送邮件
func (d *SMTPDriver) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", d.config.Host, d.config.Port)

	// 构建邮件头
	from := d.config.FromAddress
	if d.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", d.config.FromName, d.config.FromAddress)
	}

	headers := map[string]string{
		"From":         from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=UTF-8",
	}

	var message strings.Builder
	for key, value := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	auth := smtp.PlainAuth("", d.config.Username, d.config.Password, d.config.Host)

	return smtp.SendMail(addr, auth, d.config.FromAddress, []string{to}, []byte(message.String()))
}

// APIConfig API 配置
type APIConfig struct {
	Provider string // sendgrid, mailgun
	APIKey   string
	Domain   string // mailgun 需要
	From     string
}

// SendGridDriver SendGrid 邮件驱动
type SendGridDriver struct {
	config *APIConfig
}

// NewSendGridDriver 创建 SendGrid 驱动
func NewSendGridDriver(config *APIConfig) *SendGridDriver {
	return &SendGridDriver{
		config: config,
	}
}

// Send 发送邮件（SendGrid）
func (d *SendGridDriver) Send(to, subject, body string) error {
	// TODO: 实现 SendGrid API 调用
	// 这里只是示例，实际需要使用 SendGrid SDK
	return fmt.Errorf("sendgrid driver not implemented yet")
}

// MailgunDriver Mailgun 邮件驱动
type MailgunDriver struct {
	config *APIConfig
}

// NewMailgunDriver 创建 Mailgun 驱动
func NewMailgunDriver(config *APIConfig) *MailgunDriver {
	return &MailgunDriver{
		config: config,
	}
}

// Send 发送邮件（Mailgun）
func (d *MailgunDriver) Send(to, subject, body string) error {
	// TODO: 实现 Mailgun API 调用
	// 这里只是示例，实际需要使用 Mailgun SDK
	return fmt.Errorf("mailgun driver not implemented yet")
}

// EmailService 邮件服务
type EmailService struct {
	driver EmailDriver
}

// NewEmailService 创建邮件服务
func NewEmailService(driver EmailDriver) *EmailService {
	return &EmailService{
		driver: driver,
	}
}

// Send 发送邮件
func (s *EmailService) Send(to, subject, body string) error {
	return s.driver.Send(to, subject, body)
}

// SendBatch 批量发送邮件
func (s *EmailService) SendBatch(to []string, subject, body string) error {
	for _, addr := range to {
		if err := s.driver.Send(addr, subject, body); err != nil {
			return fmt.Errorf("failed to send to %s: %w", addr, err)
		}
	}
	return nil
}
