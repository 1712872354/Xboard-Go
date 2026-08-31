package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// MailLogRepository 邮件日志仓储接口
type MailLogRepository interface {
	Create(log *model.MailLog) error
	GetByID(id uint) (*model.MailLog, error)
	Update(log *model.MailLog) error
	Delete(id uint) error
	List(page, pageSize int, status int) ([]model.MailLog, int64, error)
	ListByUser(userID uint, page, pageSize int) ([]model.MailLog, int64, error)
	GetPending() ([]model.MailLog, error)
}

type mailLogRepository struct {
	db *gorm.DB
}

// NewMailLogRepository 创建邮件日志仓储
func NewMailLogRepository() MailLogRepository {
	return &mailLogRepository{
		db: database.Get(),
	}
}

// Create 创建邮件日志
func (r *mailLogRepository) Create(log *model.MailLog) error {
	return r.db.Create(log).Error
}

// GetByID 根据ID获取邮件日志
func (r *mailLogRepository) GetByID(id uint) (*model.MailLog, error) {
	var log model.MailLog
	err := r.db.First(&log, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

// Update 更新邮件日志
func (r *mailLogRepository) Update(log *model.MailLog) error {
	return r.db.Save(log).Error
}

// Delete 删除邮件日志
func (r *mailLogRepository) Delete(id uint) error {
	return r.db.Delete(&model.MailLog{}, id).Error
}

// List 邮件日志列表
func (r *mailLogRepository) List(page, pageSize int, status int) ([]model.MailLog, int64, error) {
	var logs []model.MailLog
	var total int64

	query := r.db.Model(&model.MailLog{})

	if status >= 0 {
		query = query.Where("status = ?", status)
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

// ListByUser 用户邮件日志列表
func (r *mailLogRepository) ListByUser(userID uint, page, pageSize int) ([]model.MailLog, int64, error) {
	var logs []model.MailLog
	var total int64

	query := r.db.Model(&model.MailLog{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetPending 获取待发送的邮件
func (r *mailLogRepository) GetPending() ([]model.MailLog, error) {
	var logs []model.MailLog
	err := r.db.Where("status = ?", model.MailStatusPending).Order("id ASC").Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}
