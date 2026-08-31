package tasks

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"xboard-go/internal/model"
)

// CleanupTasks 清理相关定时任务
type CleanupTasks struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewCleanupTasks 创建清理任务实例
func NewCleanupTasks(db *gorm.DB, logger *zap.Logger) *CleanupTasks {
	return &CleanupTasks{
		db:     db,
		logger: logger,
	}
}

// CleanOnlineStatus 清理过期的在线状态（每5分钟执行）
func (t *CleanupTasks) CleanOnlineStatus(ctx context.Context) error {
	// 清理超过10分钟未活跃的用户在线状态
	threshold := time.Now().Add(-10 * time.Minute)

	result := t.db.Model(&model.User{}).
		Where("online_count > 0 AND (last_online_at IS NULL OR last_online_at < ?)", threshold).
		Update("online_count", 0)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		t.logger.Info("Cleaned online status", zap.Int64("count", result.RowsAffected))
	}

	return nil
}

// CleanExpiredTickets 清理长期未关闭的工单提醒（每日执行）
func (t *CleanupTasks) CleanExpiredTickets(ctx context.Context) error {
	// 将超过30天未回复的已关闭工单标记为归档
	threshold := time.Now().AddDate(0, 0, -30)

	result := t.db.Model(&model.Ticket{}).
		Where("status = ? AND updated_at < ?", model.TicketStatusClosed, threshold).
		Update("status", model.TicketStatusArchived)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		t.logger.Info("Archived old tickets", zap.Int64("count", result.RowsAffected))
	}

	return nil
}

// CleanTrafficCache 清理过期的 Redis 流量缓存（每日执行）
func (t *CleanupTasks) CleanTrafficCache(ctx context.Context) error {
	// 这个任务会清理超过24小时的流量同步数据
	// Redis 的 TTL 机制已经自动处理，这里主要做统计
	t.logger.Debug("Traffic cache cleanup completed (Redis TTL handles expiration)")
	return nil
}
