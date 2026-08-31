package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// AdminAuditLogRepository 管理员审计日志仓储接口
type AdminAuditLogRepository interface {
	Create(log *model.AdminAuditLog) error
	GetByID(id uint) (*model.AdminAuditLog, error)
	List(page, pageSize int, userID uint, action string) ([]model.AdminAuditLog, int64, error)
	Delete(id uint) error
}

type adminAuditLogRepository struct {
	db *gorm.DB
}

// NewAdminAuditLogRepository 创建管理员审计日志仓储
func NewAdminAuditLogRepository() AdminAuditLogRepository {
	return &adminAuditLogRepository{
		db: database.Get(),
	}
}

// Create 创建审计日志
func (r *adminAuditLogRepository) Create(log *model.AdminAuditLog) error {
	return r.db.Create(log).Error
}

// GetByID 根据ID获取审计日志
func (r *adminAuditLogRepository) GetByID(id uint) (*model.AdminAuditLog, error) {
	var log model.AdminAuditLog
	err := r.db.First(&log, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

// List 审计日志列表
func (r *adminAuditLogRepository) List(page, pageSize int, userID uint, action string) ([]model.AdminAuditLog, int64, error) {
	var logs []model.AdminAuditLog
	var total int64

	query := r.db.Model(&model.AdminAuditLog{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if action != "" {
		query = query.Where("action = ?", action)
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

// Delete 删除审计日志
func (r *adminAuditLogRepository) Delete(id uint) error {
	return r.db.Delete(&model.AdminAuditLog{}, id).Error
}
