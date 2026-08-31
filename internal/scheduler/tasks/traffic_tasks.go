package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"xboard-go/internal/model"
)

// TrafficTasks 流量相关定时任务
type TrafficTasks struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewTrafficTasks 创建流量任务实例
func NewTrafficTasks(db *gorm.DB, logger *zap.Logger) *TrafficTasks {
	return &TrafficTasks{
		db:     db,
		logger: logger,
	}
}

// ResetTraffic 重置用户流量
func (t *TrafficTasks) ResetTraffic(ctx context.Context) error {
	now := time.Now()

	// 查找需要重置流量的用户（已过期的套餐）
	var users []model.User
	if err := t.db.Where("expired_at IS NOT NULL AND expired_at < ? AND status = ?", now, 1).
		Find(&users).Error; err != nil {
		return fmt.Errorf("failed to query expired users: %w", err)
	}

	if len(users) == 0 {
		return nil
	}

	t.logger.Info("Found users with expired traffic", zap.Int("count", len(users)))

	// 重置流量
	for _, user := range users {
		// 检查是否有未支付的订单（如果有，不重置）
		var pendingOrders int64
		if err := t.db.Model(&model.Order{}).
			Where("user_id = ? AND status = ?", user.ID, model.OrderStatusPending).
			Count(&pendingOrders).Error; err != nil {
			t.logger.Error("Failed to check pending orders",
				zap.Uint("user_id", user.ID),
				zap.Error(err),
			)
			continue
		}

		if pendingOrders > 0 {
			t.logger.Info("User has pending orders, skipping traffic reset",
				zap.Uint("user_id", user.ID),
			)
			continue
		}

		// 重置流量
		if err := t.db.Model(&user).Updates(map[string]interface{}{
			"used_traffic": 0,
			"expired_at":   nil,
		}).Error; err != nil {
			t.logger.Error("Failed to reset user traffic",
				zap.Uint("user_id", user.ID),
				zap.Error(err),
			)
			continue
		}

		t.logger.Info("User traffic reset",
			zap.Uint("user_id", user.ID),
			zap.String("email", user.Email),
		)
	}

	return nil
}

// CheckTrafficExceeded 检查流量超限用户
func (t *TrafficTasks) CheckTrafficExceeded(ctx context.Context) error {
	// 查找流量超限的用户
	var users []model.User
	if err := t.db.Where("traffic_limit > 0 AND used_traffic >= traffic_limit AND status = ?", 1).
		Find(&users).Error; err != nil {
		return fmt.Errorf("failed to query traffic exceeded users: %w", err)
	}

	if len(users) == 0 {
		return nil
	}

	t.logger.Info("Found users with exceeded traffic", zap.Int("count", len(users)))

	// 记录超限用户（不自动禁用，由管理员决定）
	for _, user := range users {
		t.logger.Warn("User traffic exceeded",
			zap.Uint("user_id", user.ID),
			zap.String("email", user.Email),
			zap.Int64("traffic_limit", user.TrafficLimit),
			zap.Int64("used_traffic", user.UsedTraffic),
		)
	}

	return nil
}

// GenerateTrafficStats 生成流量统计
func (t *TrafficTasks) GenerateTrafficStats(ctx context.Context) error {
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	// 统计昨日流量使用情况
	var stats struct {
		TotalUsers    int64
		ActiveUsers   int64
		TotalUpload   int64
		TotalDownload int64
	}

	// 总用户数
	if err := t.db.Model(&model.User{}).Where("status = ?", 1).Count(&stats.TotalUsers).Error; err != nil {
		return fmt.Errorf("failed to count total users: %w", err)
	}

	// 活跃用户数（有流量记录的用户）
	if err := t.db.Model(&model.TrafficLog{}).
		Where("recorded_at >= ? AND recorded_at < ?", yesterday, today).
		Distinct("user_id").
		Count(&stats.ActiveUsers).Error; err != nil {
		// 如果TrafficLog表不存在，忽略错误
		stats.ActiveUsers = 0
	}

	// 总上传流量
	if err := t.db.Model(&model.User{}).
		Where("status = ?", 1).
		Select("COALESCE(SUM(used_traffic), 0)").
		Scan(&stats.TotalUpload).Error; err != nil {
		return fmt.Errorf("failed to sum total upload: %w", err)
	}

	t.logger.Info("Daily traffic statistics",
		zap.String("date", yesterday.Format("2006-01-02")),
		zap.Int64("total_users", stats.TotalUsers),
		zap.Int64("active_users", stats.ActiveUsers),
		zap.Int64("total_upload", stats.TotalUpload),
	)

	return nil
}
