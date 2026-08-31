package tasks

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"xboard-go/internal/model"
)

// TicketTasks 工单相关定时任务
type TicketTasks struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewTicketTasks 创建工单任务实例
func NewTicketTasks(db *gorm.DB, logger *zap.Logger) *TicketTasks {
	return &TicketTasks{
		db:     db,
		logger: logger,
	}
}

// CheckPendingTickets 检查待处理工单（每分钟执行）
func (t *TicketTasks) CheckPendingTickets(ctx context.Context) error {
	// 检查超过2小时未回复的工单
	threshold := time.Now().Add(-2 * time.Hour)

	var count int64
	if err := t.db.Model(&model.Ticket{}).
		Where("status = ? AND created_at < ?", model.TicketStatusOpen, threshold).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		t.logger.Warn("Found unanswered tickets",
			zap.Int64("count", count),
			zap.Duration("threshold", 2*time.Hour),
		)
		// TODO: 发送 Telegram/邮件通知管理员
	}

	return nil
}

// CheckStaleReplies 检查用户未回复的工单（每5分钟执行）
func (t *TicketTasks) CheckStaleReplies(ctx context.Context) error {
	// 检查管理员回复后超过24小时用户未回复的工单
	threshold := time.Now().Add(-24 * time.Hour)

	var tickets []model.Ticket
	if err := t.db.Where("status = ? AND last_reply < ?",
		model.TicketStatusReplied, threshold).
		Limit(100).
		Find(&tickets).Error; err != nil {
		return err
	}

	if len(tickets) > 0 {
		t.logger.Info("Found stale ticket replies",
			zap.Int("count", len(tickets)),
		)
		// 这些工单可以自动关闭或发送提醒
	}

	return nil
}
