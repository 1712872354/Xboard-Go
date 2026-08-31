package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// NoticeRepository 公告仓储接口
type NoticeRepository interface {
	Create(notice *model.Notice) error
	GetByID(id uint) (*model.Notice, error)
	Update(notice *model.Notice) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.Notice, int64, error)
	ListVisible() ([]model.Notice, error)
}

type noticeRepository struct {
	db *gorm.DB
}

// NewNoticeRepository 创建公告仓储
func NewNoticeRepository() NoticeRepository {
	return &noticeRepository{
		db: database.Get(),
	}
}

// Create 创建公告
func (r *noticeRepository) Create(notice *model.Notice) error {
	return r.db.Create(notice).Error
}

// GetByID 根据ID获取公告
func (r *noticeRepository) GetByID(id uint) (*model.Notice, error) {
	var notice model.Notice
	err := r.db.First(&notice, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &notice, nil
}

// Update 更新公告
func (r *noticeRepository) Update(notice *model.Notice) error {
	return r.db.Save(notice).Error
}

// Delete 删除公告
func (r *noticeRepository) Delete(id uint) error {
	return r.db.Delete(&model.Notice{}, id).Error
}

// List 公告列表（管理员）
func (r *noticeRepository) List(page, pageSize int) ([]model.Notice, int64, error) {
	var notices []model.Notice
	var total int64

	query := r.db.Model(&model.Notice{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort ASC, id DESC").Find(&notices).Error; err != nil {
		return nil, 0, err
	}

	return notices, total, nil
}

// ListVisible 获取可见公告列表（用户端）
func (r *noticeRepository) ListVisible() ([]model.Notice, error) {
	var notices []model.Notice
	err := r.db.Where("show = ?", 1).Order("sort ASC, id DESC").Find(&notices).Error
	if err != nil {
		return nil, err
	}
	return notices, nil
}
