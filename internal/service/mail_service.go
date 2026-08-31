package service

import (
	"fmt"
	"regexp"
	"time"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/email"
	"xboard-go/pkg/logger"

	"gorm.io/gorm"
)

// MailService 邮件服务接口
type MailService interface {
	// SendEmail 发送邮件
	SendEmail(to, subject, templateName string, templateValue map[string]interface{}) error
	// ShouldSendExpireRemind 检查是否应该发送过期提醒
	ShouldSendExpireRemind(user *model.User) bool
	// ShouldSendTrafficRemind 检查是否应该发送流量提醒
	ShouldSendTrafficRemind(user *model.User) bool
	// GetTotalUsersNeedRemind 获取需要提醒的用户数
	GetTotalUsersNeedRemind() (int64, error)
	// ProcessUsersInChunks 分批处理用户提醒邮件
	ProcessUsersInChunks(chunkSize int) (*MailStats, error)
}

// MailStats 邮件发送统计
type MailStats struct {
	ProcessedUsers int `json:"processed_users"`
	ExpireEmails   int `json:"expire_emails"`
	TrafficEmails  int `json:"traffic_emails"`
	Errors         int `json:"errors"`
	Skipped        int `json:"skipped"`
}

type mailService struct {
	db     *gorm.DB
	driver email.EmailDriver
}

// NewMailService 创建邮件服务
func NewMailService() MailService {
	return &mailService{
		db: database.Get(),
	}
}

// getDriver 获取邮件驱动（延迟初始化）
func (s *mailService) getDriver() (email.EmailDriver, error) {
	if s.driver != nil {
		return s.driver, nil
	}

	// 从 settings 读取 SMTP 配置
	cfg := &email.SMTPConfig{}
	cfg.Host = s.getSetting("email_host", "")
	cfg.Port = 465 // 默认端口
	if portStr := s.getSetting("email_port", "465"); portStr != "" {
		fmt.Sscanf(portStr, "%d", &cfg.Port)
	}
	cfg.Username = s.getSetting("email_username", "")
	cfg.Password = s.getSetting("email_password", "")
	cfg.FromAddress = s.getSetting("email_from_address", "")
	cfg.FromName = s.getSetting("app_name", "XBoard")
	cfg.Encryption = s.getSetting("email_encryption", "ssl")

	if cfg.Host == "" || cfg.Username == "" {
		return nil, fmt.Errorf("email SMTP not configured")
	}

	s.driver = email.NewSMTPDriver(cfg)
	return s.driver, nil
}

// getSetting 从数据库获取设置值
func (s *mailService) getSetting(key, defaultValue string) string {
	var setting model.Setting
	if err := s.db.Where("`key` = ?", key).First(&setting).Error; err != nil {
		return defaultValue
	}
	if setting.Value == "" {
		return defaultValue
	}
	return setting.Value
}

// SendEmail 发送邮件
func (s *mailService) SendEmail(to, subject, templateName string, templateValue map[string]interface{}) error {
	driver, err := s.getDriver()
	if err != nil {
		return fmt.Errorf("email driver error: %w", err)
	}

	// 尝试从数据库获取模板
	var tpl model.MailTemplate
	if err := s.db.Where("name = ?", templateName).First(&tpl).Error; err == nil {
		// 使用数据库模板
		subject = renderPlaceholders(tpl.Subject, templateValue)
		body := renderPlaceholders(tpl.Body, templateValue)
		return driver.Send(to, subject, body)
	}

	// 使用内置模板
	body := s.renderBuiltinTemplate(templateName, templateValue)
	return driver.Send(to, subject, body)
}

// renderPlaceholders 渲染模板变量 {{key}} / {{key|default}}
func renderPlaceholders(template string, vars map[string]interface{}) string {
	if template == "" || len(vars) == 0 {
		return template
	}

	re := regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.-]+)(?:\|([^}]*))?\s*\}\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		matches := re.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		key := matches[1]
		defaultVal := ""
		if len(matches) > 2 {
			defaultVal = matches[2]
		}

		if val, ok := vars[key]; ok && val != nil {
			return fmt.Sprintf("%v", val)
		}
		if defaultVal != "" {
			return defaultVal
		}
		return match
	})
}

// renderBuiltinTemplate 渲染内置模板
func (s *mailService) renderBuiltinTemplate(templateName string, vars map[string]interface{}) string {
	appName := s.getSetting("app_name", "XBoard")
	appURL := s.getSetting("app_url", "")

	switch templateName {
	case "remindExpire":
		return fmt.Sprintf(`<h2>%s - 服务到期提醒</h2>
<p>尊敬的用户您好：</p>
<p>您的服务将在 <strong>24小时</strong> 内到期，请及时续费以避免影响使用。</p>
<p><a href="%s">点击此处登录续费</a></p>
<p>感谢您的支持！</p>
<p>%s 团队</p>`, appName, appURL, appName)

	case "remindTraffic":
		return fmt.Sprintf(`<h2>%s - 流量使用提醒</h2>
<p>尊敬的用户您好：</p>
<p>您的流量使用已超过 <strong>80%%</strong>，请注意合理使用。</p>
<p><a href="%s">点击此处查看详情</a></p>
<p>感谢您的支持！</p>
<p>%s 团队</p>`, appName, appURL, appName)

	case "notify":
		content := vars["content"]
		return fmt.Sprintf(`<h2>%s</h2>
<p>%v</p>
<p><a href="%s">%s</a></p>`, appName, content, appURL, appName)

	default:
		return fmt.Sprintf("<p>%s notification</p>", appName)
	}
}

// ShouldSendExpireRemind 检查是否应该发送过期提醒
func (s *mailService) ShouldSendExpireRemind(user *model.User) bool {
	if user.ExpiredAt == nil {
		return false
	}
	now := time.Now()
	// 到期前24小时内发送
	return user.ExpiredAt.Sub(now) > 0 && user.ExpiredAt.Sub(now) <= 24*time.Hour
}

// ShouldSendTrafficRemind 检查是否应该发送流量提醒
func (s *mailService) ShouldSendTrafficRemind(user *model.User) bool {
	if user.TrafficLimit <= 0 {
		return false
	}
	usageRatio := float64(user.UsedTraffic) / float64(user.TrafficLimit)
	return usageRatio >= 0.8 && usageRatio < 1.0
}

// GetTotalUsersNeedRemind 获取需要提醒的用户数
func (s *mailService) GetTotalUsersNeedRemind() (int64, error) {
	var count int64
	err := s.db.Model(&model.User{}).
		Where("(remind_expire = ? OR remind_traffic = ?) AND status = ? AND email != ''",
			true, true, 1).
		Count(&count).Error
	return count, err
}

// ProcessUsersInChunks 分批处理用户提醒邮件
func (s *mailService) ProcessUsersInChunks(chunkSize int) (*MailStats, error) {
	if chunkSize <= 0 {
		chunkSize = 100
	}

	stats := &MailStats{}
	var lastID uint

	for {
		var users []model.User
		err := s.db.Where("id > ? AND (remind_expire = ? OR remind_traffic = ?) AND status = ? AND email != ''",
			lastID, true, true, 1).
			Order("id ASC").
			Limit(chunkSize).
			Find(&users).Error

		if err != nil {
			return stats, fmt.Errorf("failed to query users: %w", err)
		}

		if len(users) == 0 {
			break
		}

		for i := range users {
			stats.ProcessedUsers++
			emailsSent := 0

			// 检查过期提醒
			if users[i].RemindExpire && s.ShouldSendExpireRemind(&users[i]) {
				if err := s.SendEmail(users[i].Email,
					fmt.Sprintf("%s - 服务到期提醒", s.getSetting("app_name", "XBoard")),
					"remindExpire",
					map[string]interface{}{
						"name": s.getSetting("app_name", "XBoard"),
						"url":  s.getSetting("app_url", ""),
					}); err != nil {
					logger.Sugar().Warnf("Failed to send expire remind to %s: %v", users[i].Email, err)
					stats.Errors++
				} else {
					stats.ExpireEmails++
					emailsSent++
				}
			}

			// 检查流量提醒
			if users[i].RemindTraffic && s.ShouldSendTrafficRemind(&users[i]) {
				if err := s.SendEmail(users[i].Email,
					fmt.Sprintf("%s - 流量使用提醒", s.getSetting("app_name", "XBoard")),
					"remindTraffic",
					map[string]interface{}{
						"name": s.getSetting("app_name", "XBoard"),
						"url":  s.getSetting("app_url", ""),
					}); err != nil {
					logger.Sugar().Warnf("Failed to send traffic remind to %s: %v", users[i].Email, err)
					stats.Errors++
				} else {
					stats.TrafficEmails++
					emailsSent++
				}
			}

			if emailsSent == 0 {
				stats.Skipped++
			}

			lastID = users[i].ID
		}

		// 定期清理内存
		if stats.ProcessedUsers%1000 == 0 {
			logger.Sugar().Debugf("Mail remind progress: processed=%d", stats.ProcessedUsers)
		}
	}

	return stats, nil
}
