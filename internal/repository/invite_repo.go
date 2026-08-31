package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// InviteCodeRepository 邀请码仓储接口
type InviteCodeRepository interface {
	Create(code *model.InviteCode) error
	GetByID(id uint) (*model.InviteCode, error)
	GetByCode(code string) (*model.InviteCode, error)
	GetByUserID(userID uint) (*model.InviteCode, error)
	Update(code *model.InviteCode) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.InviteCode, int64, error)
	IncrementUsedCount(id uint) error
}

type inviteCodeRepository struct {
	db *gorm.DB
}

// NewInviteCodeRepository 创建邀请码仓储
func NewInviteCodeRepository() InviteCodeRepository {
	return &inviteCodeRepository{
		db: database.Get(),
	}
}

// Create 创建邀请码
func (r *inviteCodeRepository) Create(code *model.InviteCode) error {
	return r.db.Create(code).Error
}

// GetByID 根据ID获取邀请码
func (r *inviteCodeRepository) GetByID(id uint) (*model.InviteCode, error) {
	var code model.InviteCode
	err := r.db.First(&code, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &code, nil
}

// GetByCode 根据邀请码获取
func (r *inviteCodeRepository) GetByCode(code string) (*model.InviteCode, error) {
	var inviteCode model.InviteCode
	err := r.db.Where("code = ?", code).First(&inviteCode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &inviteCode, nil
}

// GetByUserID 根据用户ID获取邀请码
func (r *inviteCodeRepository) GetByUserID(userID uint) (*model.InviteCode, error) {
	var code model.InviteCode
	err := r.db.Where("user_id = ?", userID).First(&code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &code, nil
}

// Update 更新邀请码
func (r *inviteCodeRepository) Update(code *model.InviteCode) error {
	return r.db.Save(code).Error
}

// Delete 删除邀请码
func (r *inviteCodeRepository) Delete(id uint) error {
	return r.db.Delete(&model.InviteCode{}, id).Error
}

// List 邀请码列表
func (r *inviteCodeRepository) List(page, pageSize int) ([]model.InviteCode, int64, error) {
	var codes []model.InviteCode
	var total int64

	query := r.db.Model(&model.InviteCode{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&codes).Error; err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}

// IncrementUsedCount 增加使用次数
func (r *inviteCodeRepository) IncrementUsedCount(id uint) error {
	return r.db.Model(&model.InviteCode{}).Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}

// CommissionLogRepository 佣金记录仓储接口
type CommissionLogRepository interface {
	Create(log *model.CommissionLog) error
	GetByID(id uint) (*model.CommissionLog, error)
	Update(log *model.CommissionLog) error
	List(page, pageSize int, userID uint) ([]model.CommissionLog, int64, error)
	GetTotalByUser(userID uint) (float64, error)
	GetPendingByUser(userID uint) (float64, error)
}

type commissionLogRepository struct {
	db *gorm.DB
}

// NewCommissionLogRepository 创建佣金记录仓储
func NewCommissionLogRepository() CommissionLogRepository {
	return &commissionLogRepository{
		db: database.Get(),
	}
}

// Create 创建佣金记录
func (r *commissionLogRepository) Create(log *model.CommissionLog) error {
	return r.db.Create(log).Error
}

// GetByID 根据ID获取佣金记录
func (r *commissionLogRepository) GetByID(id uint) (*model.CommissionLog, error) {
	var log model.CommissionLog
	err := r.db.First(&log, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

// Update 更新佣金记录
func (r *commissionLogRepository) Update(log *model.CommissionLog) error {
	return r.db.Save(log).Error
}

// List 佣金记录列表
func (r *commissionLogRepository) List(page, pageSize int, userID uint) ([]model.CommissionLog, int64, error) {
	var logs []model.CommissionLog
	var total int64

	query := r.db.Model(&model.CommissionLog{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
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

// GetTotalByUser 获取用户总佣金
func (r *commissionLogRepository) GetTotalByUser(userID uint) (float64, error) {
	var total float64
	err := r.db.Model(&model.CommissionLog{}).
		Where("user_id = ? AND status = ?", userID, model.CommissionStatusSettled).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

// GetPendingByUser 获取用户待结算佣金
func (r *commissionLogRepository) GetPendingByUser(userID uint) (float64, error) {
	var total float64
	err := r.db.Model(&model.CommissionLog{}).
		Where("user_id = ? AND status = ?", userID, model.CommissionStatusPending).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}
