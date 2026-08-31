package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"xboard-go/internal/model"
	"xboard-go/internal/service"
)

// TrafficResetTasks 流量重置定时任务
type TrafficResetTasks struct {
	db            *gorm.DB
	logger        *zap.Logger
	resetService  service.TrafficResetService
}

// NewTrafficResetTasks 创建流量重置任务实例
func NewTrafficResetTasks(db *gorm.DB, logger *zap.Logger) *TrafficResetTasks {
	return &TrafficResetTasks{
		db:           db,
		logger:       logger,
		resetService: service.NewTrafficResetService(),
	}
}

// CheckAndResetTraffic 检查并重置用户流量（每分钟执行）
func (t *TrafficResetTasks) CheckAndResetTraffic(ctx context.Context) error {
	totalProcessed, totalReset, err := t.resetService.BatchCheckReset(100)
	if err != nil {
		return fmt.Errorf("batch reset failed: %w", err)
	}

	if totalReset > 0 {
		t.logger.Info("Traffic reset completed",
			zap.Int("processed", totalProcessed),
			zap.Int("reset", totalReset),
		)
	}

	return nil
}

// SetInitialResetTimes 为缺少重置时间的用户设置初始值（每日执行一次）
func (t *TrafficResetTasks) SetInitialResetTimes(ctx context.Context) error {
	var users []model.User
	// 查找有套餐但没有设置重置时间的用户
	if err := t.db.Where("plan_id IS NOT NULL AND next_reset_at IS NULL AND status = ?", 1).
		Where("expired_at IS NULL OR expired_at > ?", time.Now()).
		Limit(500).
		Find(&users).Error; err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}

	if len(users) == 0 {
		return nil
	}

	count := 0
	for i := range users {
		var plan model.Plan
		if users[i].PlanID == nil {
			continue
		}
		if err := t.db.First(&plan, *users[i].PlanID).Error; err != nil {
			continue
		}

		t.resetService.SetInitialResetTime(&users[i], &plan)
		count++
	}

	if count > 0 {
		t.logger.Info("Set initial reset times", zap.Int("count", count))
	}

	return nil
}
