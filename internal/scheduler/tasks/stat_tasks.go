package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"xboard-go/internal/model"
)

// StatTasks 统计相关定时任务
type StatTasks struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewStatTasks 创建统计任务实例
func NewStatTasks(db *gorm.DB, logger *zap.Logger) *StatTasks {
	return &StatTasks{
		db:     db,
		logger: logger,
	}
}

// GenerateDailyStats 生成每日综合统计（建议每日 00:10 执行）
func (t *StatTasks) GenerateDailyStats(ctx context.Context) error {
	yesterday := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -1)
	today := yesterday.AddDate(0, 0, 1)

	// 检查是否已经生成过
	var existing model.Stat
	if err := t.db.Where("record_at >= ? AND record_at < ? AND record_type = ?",
		yesterday, today, "d").First(&existing).Error; err == nil {
		t.logger.Debug("Daily stats already generated", zap.String("date", yesterday.Format("2006-01-02")))
		return nil
	}

	stat := model.Stat{
		RecordAt:   yesterday,
		RecordType: "d",
	}

	// 订单统计
	t.db.Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ?", yesterday, today).
		Count(&stat.OrderCount)

	t.db.Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ? AND status != ? AND status != ?",
			yesterday, today, model.OrderStatusPending, model.OrderStatusCancelled).
		Count(&stat.PaidCount)

	t.db.Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ?", yesterday, today).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stat.OrderTotal)

	t.db.Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ? AND status != ? AND status != ?",
			yesterday, today, model.OrderStatusPending, model.OrderStatusCancelled).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stat.PaidTotal)

	// 注册统计
	t.db.Model(&model.User{}).
		Where("created_at >= ? AND created_at < ?", yesterday, today).
		Count(&stat.RegisterCount)

	// 邀请注册统计
	t.db.Model(&model.User{}).
		Where("created_at >= ? AND created_at < ? AND invite_code_id IS NOT NULL",
			yesterday, today).
		Count(&stat.InviteCount)

	// 佣金统计
	t.db.Model(&model.CommissionLog{}).
		Where("created_at >= ? AND created_at < ?", yesterday, today).
		Count(&stat.CommissionCount)

	t.db.Model(&model.CommissionLog{}).
		Where("created_at >= ? AND created_at < ?", yesterday, today).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stat.CommissionTotal)

	// 流量统计
	t.db.Model(&model.TrafficLog{}).
		Where("recorded_at >= ? AND recorded_at < ?", yesterday, today).
		Select("COALESCE(SUM(upload) + SUM(download), 0)").
		Scan(&stat.TransferUsed)

	// 保存统计记录
	if err := t.db.Create(&stat).Error; err != nil {
		return fmt.Errorf("failed to create daily stat: %w", err)
	}

	t.logger.Info("Daily stats generated",
		zap.String("date", yesterday.Format("2006-01-02")),
		zap.Int64("orders", stat.OrderCount),
		zap.Int64("paid", stat.PaidCount),
		zap.Int64("registers", stat.RegisterCount),
	)

	return nil
}

// GenerateDailyServerStats 生成每日节点流量统计
func (t *StatTasks) GenerateDailyServerStats(ctx context.Context) error {
	yesterday := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -1)
	today := yesterday.AddDate(0, 0, 1)

	// 检查是否已经生成过
	var count int64
	t.db.Model(&model.StatServer{}).Where("record_at >= ? AND record_at < ?", yesterday, today).Count(&count)
	if count > 0 {
		t.logger.Debug("Daily server stats already generated", zap.String("date", yesterday.Format("2006-01-02")))
		return nil
	}

	// 从 TrafficLog 聚合节点流量
	type NodeTraffic struct {
		NodeID   uint
		Type     string
		Upload   int64
		Download int64
	}

	var results []NodeTraffic
	t.db.Model(&model.TrafficLog{}).
		Select("node_id, '' as type, COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download").
		Where("recorded_at >= ? AND recorded_at < ?", yesterday, today).
		Group("node_id").
		Scan(&results)

	if len(results) == 0 {
		return nil
	}

	// 获取节点类型信息
	nodeIDs := make([]uint, len(results))
	for i, r := range results {
		nodeIDs[i] = r.NodeID
	}
	var nodes []model.Node
	t.db.Where("id IN ?", nodeIDs).Find(&nodes)
	nodeTypeMap := make(map[uint]string)
	for _, n := range nodes {
		nodeTypeMap[n.ID] = n.Type
	}

	// 批量插入
	var stats []model.StatServer
	for _, r := range results {
		stats = append(stats, model.StatServer{
			ServerID:   r.NodeID,
			ServerType: nodeTypeMap[r.NodeID],
			Upload:     r.Upload,
			Download:   r.Download,
			RecordAt:   yesterday,
		})
	}

	if err := t.db.CreateInBatches(stats, 100).Error; err != nil {
		return fmt.Errorf("failed to create server stats: %w", err)
	}

	t.logger.Info("Daily server stats generated",
		zap.String("date", yesterday.Format("2006-01-02")),
		zap.Int("nodes", len(stats)),
	)

	return nil
}

// GenerateDailyUserStats 生成每日用户流量统计
func (t *StatTasks) GenerateDailyUserStats(ctx context.Context) error {
	yesterday := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -1)
	today := yesterday.AddDate(0, 0, 1)

	// 检查是否已经生成过
	var count int64
	t.db.Model(&model.StatUser{}).Where("record_at >= ? AND record_at < ?", yesterday, today).Count(&count)
	if count > 0 {
		t.logger.Debug("Daily user stats already generated", zap.String("date", yesterday.Format("2006-01-02")))
		return nil
	}

	// 从 TrafficLog 聚合用户流量
	type UserTraffic struct {
		UserID   uint
		Upload   int64
		Download int64
	}

	var results []UserTraffic
	t.db.Model(&model.TrafficLog{}).
		Select("user_id, COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download").
		Where("recorded_at >= ? AND recorded_at < ?", yesterday, today).
		Group("user_id").
		Scan(&results)

	if len(results) == 0 {
		return nil
	}

	// 批量插入
	var stats []model.StatUser
	for _, r := range results {
		stats = append(stats, model.StatUser{
			UserID:     r.UserID,
			ServerRate: 1,
			Upload:     r.Upload,
			Download:   r.Download,
			RecordAt:   yesterday,
		})
	}

	if err := t.db.CreateInBatches(stats, 100).Error; err != nil {
		return fmt.Errorf("failed to create user stats: %w", err)
	}

	t.logger.Info("Daily user stats generated",
		zap.String("date", yesterday.Format("2006-01-02")),
		zap.Int("users", len(stats)),
	)

	return nil
}
