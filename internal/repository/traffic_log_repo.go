package repository

import (
	"time"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// TrafficLogRepository 流量日志仓储接口
type TrafficLogRepository interface {
	Create(log *model.TrafficLog) error
	BatchCreate(logs []model.TrafficLog) error
	ListByUser(userID uint, start, end time.Time, page, pageSize int) ([]model.TrafficLog, int64, error)
	SumByUserAndRange(userID uint, start, end time.Time) (int64, int64, error)
}

type trafficLogRepository struct {
	db *gorm.DB
}

// NewTrafficLogRepository 创建流量日志仓储
func NewTrafficLogRepository() TrafficLogRepository {
	return &trafficLogRepository{
		db: database.Get(),
	}
}

// Create 创建流量日志
func (r *trafficLogRepository) Create(log *model.TrafficLog) error {
	return r.db.Create(log).Error
}

// BatchCreate 批量创建流量日志
func (r *trafficLogRepository) BatchCreate(logs []model.TrafficLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.Create(&logs).Error
}

// ListByUser 获取用户流量日志
func (r *trafficLogRepository) ListByUser(userID uint, start, end time.Time, page, pageSize int) ([]model.TrafficLog, int64, error) {
	var logs []model.TrafficLog
	var total int64

	query := r.db.Model(&model.TrafficLog{}).Where("user_id = ?", userID)

	if !start.IsZero() {
		query = query.Where("recorded_at >= ?", start)
	}
	if !end.IsZero() {
		query = query.Where("recorded_at <= ?", end)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("recorded_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// SumByUserAndRange 统计用户在时间段内的流量总和
func (r *trafficLogRepository) SumByUserAndRange(userID uint, start, end time.Time) (int64, int64, error) {
	var result struct {
		Upload   int64
		Download int64
	}

	query := r.db.Model(&model.TrafficLog{}).
		Select("COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download").
		Where("user_id = ?", userID)

	if !start.IsZero() {
		query = query.Where("recorded_at >= ?", start)
	}
	if !end.IsZero() {
		query = query.Where("recorded_at <= ?", end)
	}

	err := query.Scan(&result).Error
	return result.Upload, result.Download, err
}
