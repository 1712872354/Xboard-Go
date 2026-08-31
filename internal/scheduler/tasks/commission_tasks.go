package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"xboard-go/internal/model"
)

// CommissionTasks 佣金相关定时任务
type CommissionTasks struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewCommissionTasks 创建佣金任务实例
func NewCommissionTasks(db *gorm.DB, logger *zap.Logger) *CommissionTasks {
	return &CommissionTasks{
		db:     db,
		logger: logger,
	}
}

// CheckCommission 检查并结算佣金
func (t *CommissionTasks) CheckCommission(ctx context.Context) error {
	// 查找待结算的佣金记录
	var logs []model.CommissionLog
	if err := t.db.Where("status = ?", 0).Find(&logs).Error; err != nil {
		return fmt.Errorf("failed to query commission logs: %w", err)
	}

	if len(logs) == 0 {
		return nil
	}

	t.logger.Info("Found pending commission logs", zap.Int("count", len(logs)))

	// 结算佣金
	for _, log := range logs {
		// 检查关联订单是否已支付
		var order model.Order
		if err := t.db.First(&order, log.OrderID).Error; err != nil {
			t.logger.Error("Failed to find order",
				zap.Uint("commission_id", log.ID),
				zap.Uint("order_id", log.OrderID),
				zap.Error(err),
			)
			continue
		}

		if order.Status != model.OrderStatusPaid {
			t.logger.Info("Order not paid, skipping commission",
				zap.Uint("commission_id", log.ID),
				zap.Uint("order_id", log.OrderID),
			)
			continue
		}

		// 更新佣金状态为已结算
		now := time.Now()
		if err := t.db.Model(&log).Updates(map[string]interface{}{
			"status":      1,
			"settled_at":  now,
		}).Error; err != nil {
			t.logger.Error("Failed to settle commission",
				zap.Uint("commission_id", log.ID),
				zap.Error(err),
			)
			continue
		}

		// 增加用户佣金余额
		if err := t.db.Model(&model.User{}).
			Where("id = ?", log.UserID).
			Update("commission", gorm.Expr("commission + ?", log.Amount)).Error; err != nil {
			t.logger.Error("Failed to add commission to user",
				zap.Uint("user_id", log.UserID),
				zap.Float64("amount", log.Amount),
				zap.Error(err),
			)
			continue
		}

		t.logger.Info("Commission settled",
			zap.Uint("commission_id", log.ID),
			zap.Uint("user_id", log.UserID),
			zap.Float64("amount", log.Amount),
		)
	}

	return nil
}

// GenerateCommissionStats 生成佣金统计
func (t *CommissionTasks) GenerateCommissionStats(ctx context.Context) error {
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	// 统计昨日佣金数据
	var stats struct {
		TotalCommission   float64
		SettledCommission float64
		PendingCommission float64
	}

	// 总佣金
	if err := t.db.Model(&model.CommissionLog{}).
		Where("created_at >= ? AND created_at < ?", yesterday, today).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalCommission).Error; err != nil {
		return fmt.Errorf("failed to sum total commission: %w", err)
	}

	// 已结算佣金
	if err := t.db.Model(&model.CommissionLog{}).
		Where("created_at >= ? AND created_at < ? AND status = ?", yesterday, today, 1).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.SettledCommission).Error; err != nil {
		return fmt.Errorf("failed to sum settled commission: %w", err)
	}

	// 待结算佣金
	if err := t.db.Model(&model.CommissionLog{}).
		Where("created_at >= ? AND created_at < ? AND status = ?", yesterday, today, 0).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.PendingCommission).Error; err != nil {
		return fmt.Errorf("failed to sum pending commission: %w", err)
	}

	t.logger.Info("Daily commission statistics",
		zap.String("date", yesterday.Format("2006-01-02")),
		zap.Float64("total_commission", stats.TotalCommission),
		zap.Float64("settled_commission", stats.SettledCommission),
		zap.Float64("pending_commission", stats.PendingCommission),
	)

	return nil
}
