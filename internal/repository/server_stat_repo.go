package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// ServerStatRepository 服务器统计仓储接口
type ServerStatRepository interface {
	Create(stat *model.ServerStat) error
	GetByID(id uint) (*model.ServerStat, error)
	Update(stat *model.ServerStat) error
	List(page, pageSize int, serverID uint) ([]model.ServerStat, int64, error)
	GetByDate(serverID uint, date string) (*model.ServerStat, error)
	GetDateRange(serverID uint, startDate, endDate string) ([]model.ServerStat, error)
}

type serverStatRepository struct {
	db *gorm.DB
}

// NewServerStatRepository 创建服务器统计仓储
func NewServerStatRepository() ServerStatRepository {
	return &serverStatRepository{
		db: database.Get(),
	}
}

// Create 创建服务器统计
func (r *serverStatRepository) Create(stat *model.ServerStat) error {
	return r.db.Create(stat).Error
}

// GetByID 根据ID获取服务器统计
func (r *serverStatRepository) GetByID(id uint) (*model.ServerStat, error) {
	var stat model.ServerStat
	err := r.db.First(&stat, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &stat, nil
}

// Update 更新服务器统计
func (r *serverStatRepository) Update(stat *model.ServerStat) error {
	return r.db.Save(stat).Error
}

// List 服务器统计列表
func (r *serverStatRepository) List(page, pageSize int, serverID uint) ([]model.ServerStat, int64, error) {
	var stats []model.ServerStat
	var total int64

	query := r.db.Model(&model.ServerStat{})

	if serverID > 0 {
		query = query.Where("server_id = ?", serverID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("date DESC").Find(&stats).Error; err != nil {
		return nil, 0, err
	}

	return stats, total, nil
}

// GetByDate 根据日期获取服务器统计
func (r *serverStatRepository) GetByDate(serverID uint, date string) (*model.ServerStat, error) {
	var stat model.ServerStat
	err := r.db.Where("server_id = ? AND date = ?", serverID, date).First(&stat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &stat, nil
}

// GetDateRange 获取日期范围内的服务器统计
func (r *serverStatRepository) GetDateRange(serverID uint, startDate, endDate string) ([]model.ServerStat, error) {
	var stats []model.ServerStat
	err := r.db.Where("server_id = ? AND date >= ? AND date <= ?", serverID, startDate, endDate).
		Order("date ASC").Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// ServerLogRepository 服务器日志仓储接口
type ServerLogRepository interface {
	Create(log *model.ServerLog) error
	GetByID(id uint) (*model.ServerLog, error)
	List(page, pageSize int, serverID uint, level string) ([]model.ServerLog, int64, error)
	Delete(id uint) error
}

type serverLogRepository struct {
	db *gorm.DB
}

// NewServerLogRepository 创建服务器日志仓储
func NewServerLogRepository() ServerLogRepository {
	return &serverLogRepository{
		db: database.Get(),
	}
}

// Create 创建服务器日志
func (r *serverLogRepository) Create(log *model.ServerLog) error {
	return r.db.Create(log).Error
}

// GetByID 根据ID获取服务器日志
func (r *serverLogRepository) GetByID(id uint) (*model.ServerLog, error) {
	var log model.ServerLog
	err := r.db.First(&log, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

// List 服务器日志列表
func (r *serverLogRepository) List(page, pageSize int, serverID uint, level string) ([]model.ServerLog, int64, error) {
	var logs []model.ServerLog
	var total int64

	query := r.db.Model(&model.ServerLog{})

	if serverID > 0 {
		query = query.Where("server_id = ?", serverID)
	}

	if level != "" {
		query = query.Where("level = ?", level)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// Delete 删除服务器日志
func (r *serverLogRepository) Delete(id uint) error {
	return r.db.Delete(&model.ServerLog{}, id).Error
}
