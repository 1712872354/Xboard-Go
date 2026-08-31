package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"xboard-go/internal/model"
)

// OrderTasks 订单相关定时任务
type OrderTasks struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewOrderTasks 创建订单任务实例
func NewOrderTasks(db *gorm.DB, logger *zap.Logger) *OrderTasks {
	return &OrderTasks{
		db:     db,
		logger: logger,
	}
}

// CheckPendingOrders 检查待处理订单（超时自动取消）
func (t *OrderTasks) CheckPendingOrders(ctx context.Context) error {
	// 订单超时时间：30分钟
	timeout := 30 * time.Minute
	deadline := time.Now().Add(-timeout)

	// 查找超时的待支付订单
	var orders []model.Order
	if err := t.db.Where("status = ? AND created_at < ?", model.OrderStatusPending, deadline).
		Find(&orders).Error; err != nil {
		return fmt.Errorf("failed to query pending orders: %w", err)
	}

	if len(orders) == 0 {
		return nil
	}

	t.logger.Info("Found expired pending orders", zap.Int("count", len(orders)))

	// 批量更新为已取消
	for _, order := range orders {
		if err := t.db.Model(&order).Update("status", model.OrderStatusCancelled).Error; err != nil {
			t.logger.Error("Failed to cancel order",
				zap.Uint("order_id", order.ID),
				zap.Error(err),
			)
			continue
		}
		t.logger.Info("Order cancelled due to timeout",
			zap.Uint("order_id", order.ID),
			zap.String("trade_no", order.TradeNo),
		)
	}

	return nil
}

// CheckProcessingOrders 检查处理中的订单（长时间未完成）
func (t *OrderTasks) CheckProcessingOrders(ctx context.Context) error {
	// 处理中订单超时时间：2小时
	timeout := 2 * time.Hour
	deadline := time.Now().Add(-timeout)

	// 查找长时间处于处理中的订单
	var orders []model.Order
	if err := t.db.Where("status = ? AND updated_at < ?", model.OrderStatusPending, deadline).
		Find(&orders).Error; err != nil {
		return fmt.Errorf("failed to query processing orders: %w", err)
	}

	if len(orders) == 0 {
		return nil
	}

	t.logger.Info("Found stuck processing orders", zap.Int("count", len(orders)))

	// 记录日志，需要人工处理
	for _, order := range orders {
		t.logger.Warn("Order stuck in processing state",
			zap.Uint("order_id", order.ID),
			zap.String("trade_no", order.TradeNo),
			zap.Time("created_at", order.CreatedAt),
		)
	}

	return nil
}

// GenerateDailyOrderStats 生成每日订单统计
func (t *OrderTasks) GenerateDailyOrderStats(ctx context.Context) error {
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	// 统计昨日订单数据
	var stats struct {
		TotalOrders  int64
		PaidOrders   int64
		TotalAmount  float64
		PaidAmount   float64
	}

	// 订单总数
	if err := t.db.Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ?", yesterday, today).
		Count(&stats.TotalOrders).Error; err != nil {
		return fmt.Errorf("failed to count total orders: %w", err)
	}

	// 已支付订单数
	if err := t.db.Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ? AND status = ?", yesterday, today, model.OrderStatusPaid).
		Count(&stats.PaidOrders).Error; err != nil {
		return fmt.Errorf("failed to count paid orders: %w", err)
	}

	// 订单总金额
	if err := t.db.Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ?", yesterday, today).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalAmount).Error; err != nil {
		return fmt.Errorf("failed to sum total amount: %w", err)
	}

	// 已支付金额
	if err := t.db.Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ? AND status = ?", yesterday, today, model.OrderStatusPaid).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.PaidAmount).Error; err != nil {
		return fmt.Errorf("failed to sum paid amount: %w", err)
	}

	t.logger.Info("Daily order statistics",
		zap.String("date", yesterday.Format("2006-01-02")),
		zap.Int64("total_orders", stats.TotalOrders),
		zap.Int64("paid_orders", stats.PaidOrders),
		zap.Float64("total_amount", stats.TotalAmount),
		zap.Float64("paid_amount", stats.PaidAmount),
	)

	// TODO: 将统计数据保存到统计表中

	return nil
}
