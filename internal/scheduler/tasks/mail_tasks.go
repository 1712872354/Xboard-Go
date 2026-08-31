package tasks

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"xboard-go/internal/service"
)

// MailTasks 邮件相关定时任务
type MailTasks struct {
	db          *gorm.DB
	logger      *zap.Logger
	mailService service.MailService
}

// NewMailTasks 创建邮件任务实例
func NewMailTasks(db *gorm.DB, logger *zap.Logger) *MailTasks {
	return &MailTasks{
		db:          db,
		logger:      logger,
		mailService: service.NewMailService(),
	}
}

// SendRemindMails 发送提醒邮件（每日执行一次，建议 11:30）
func (t *MailTasks) SendRemindMails(ctx context.Context) error {
	// 先检查是否配置了邮件
	total, err := t.mailService.GetTotalUsersNeedRemind()
	if err != nil {
		return err
	}

	if total == 0 {
		t.logger.Debug("No users need email reminders")
		return nil
	}

	t.logger.Info("Starting email reminders", zap.Int64("total_users", total))

	stats, err := t.mailService.ProcessUsersInChunks(100)
	if err != nil {
		t.logger.Error("Mail reminder task failed", zap.Error(err))
		return err
	}

	t.logger.Info("Email reminders completed",
		zap.Int("processed", stats.ProcessedUsers),
		zap.Int("expire_emails", stats.ExpireEmails),
		zap.Int("traffic_emails", stats.TrafficEmails),
		zap.Int("skipped", stats.Skipped),
		zap.Int("errors", stats.Errors),
	)

	return nil
}
