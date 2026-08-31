package repository

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// SettingRepository 系统设置仓储接口
type SettingRepository interface {
	GetByKey(key string) (*model.Setting, error)
	GetByGroup(group string) ([]model.Setting, error)
	GetAll() ([]model.Setting, error)
	Set(key, value, group, remark string) error
	SetBatch(settings []model.Setting) error
	Delete(key string) error
}

type settingRepository struct {
	db *gorm.DB
}

// NewSettingRepository 创建系统设置仓储
func NewSettingRepository() SettingRepository {
	return &settingRepository{
		db: database.Get(),
	}
}

// GetByKey 根据键名获取设置
func (r *settingRepository) GetByKey(key string) (*model.Setting, error) {
	var setting model.Setting
	err := r.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &setting, nil
}

// GetByGroup 根据分组获取设置列表
func (r *settingRepository) GetByGroup(group string) ([]model.Setting, error) {
	var settings []model.Setting
	err := r.db.Where("`group` = ?", group).Order("id ASC").Find(&settings).Error
	if err != nil {
		return nil, err
	}
	return settings, nil
}

// GetAll 获取所有设置
func (r *settingRepository) GetAll() ([]model.Setting, error) {
	var settings []model.Setting
	err := r.db.Order("group ASC, id ASC").Find(&settings).Error
	if err != nil {
		return nil, err
	}
	return settings, nil
}

// Set 设置单个配置项（存在则更新，不存在则创建）
func (r *settingRepository) Set(key, value, group, remark string) error {
	var setting model.Setting
	err := r.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新设置
			setting = model.Setting{
				Key:    key,
				Value:  value,
				Group:  group,
				Remark: remark,
			}
			return r.db.Create(&setting).Error
		}
		return err
	}
	// 更新现有设置
	setting.Value = value
	if group != "" {
		setting.Group = group
	}
	if remark != "" {
		setting.Remark = remark
	}
	return r.db.Save(&setting).Error
}

// SetBatch 批量设置配置项
func (r *settingRepository) SetBatch(settings []model.Setting) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, setting := range settings {
			var existing model.Setting
			err := tx.Where("key = ?", setting.Key).First(&existing).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 创建新设置
					if err := tx.Create(&setting).Error; err != nil {
						return err
					}
					continue
				}
				return err
			}
			// 更新现有设置
			existing.Value = setting.Value
			if setting.Group != "" {
				existing.Group = setting.Group
			}
			if setting.Remark != "" {
				existing.Remark = setting.Remark
			}
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete 删除设置
func (r *settingRepository) Delete(key string) error {
	return r.db.Where("key = ?", key).Delete(&model.Setting{}).Error
}
