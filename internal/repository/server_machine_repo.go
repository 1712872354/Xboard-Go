package repository

import (
	"errors"
	"time"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// ServerMachineRepository 服务器机器仓储接口
type ServerMachineRepository interface {
	Create(machine *model.ServerMachine) error
	GetByID(id uint) (*model.ServerMachine, error)
	GetByToken(token string) (*model.ServerMachine, error)
	Update(machine *model.ServerMachine) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.ServerMachine, int64, error)
	ListAll() ([]model.ServerMachine, error)
	UpdateStatus(id uint, status int) error
	UpdateLoad(id uint, cpu, memory, disk float64) error
	UpdateToken(id uint, token string) error
	UpdateMachineStatus(id uint, cpu, memory, disk float64, uptime int64) error
}

type serverMachineRepository struct {
	db *gorm.DB
}

// NewServerMachineRepository 创建服务器机器仓储
func NewServerMachineRepository() ServerMachineRepository {
	return &serverMachineRepository{
		db: database.Get(),
	}
}

// Create 创建服务器机器
func (r *serverMachineRepository) Create(machine *model.ServerMachine) error {
	return r.db.Create(machine).Error
}

// GetByID 根据ID获取服务器机器
func (r *serverMachineRepository) GetByID(id uint) (*model.ServerMachine, error) {
	var machine model.ServerMachine
	err := r.db.First(&machine, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &machine, nil
}

// Update 更新服务器机器
func (r *serverMachineRepository) Update(machine *model.ServerMachine) error {
	return r.db.Save(machine).Error
}

// Delete 删除服务器机器
func (r *serverMachineRepository) Delete(id uint) error {
	return r.db.Delete(&model.ServerMachine{}, id).Error
}

// List 服务器机器列表
func (r *serverMachineRepository) List(page, pageSize int) ([]model.ServerMachine, int64, error) {
	var machines []model.ServerMachine
	var total int64

	query := r.db.Model(&model.ServerMachine{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&machines).Error; err != nil {
		return nil, 0, err
	}

	return machines, total, nil
}

// ListAll 获取所有服务器机器
func (r *serverMachineRepository) ListAll() ([]model.ServerMachine, error) {
	var machines []model.ServerMachine
	err := r.db.Order("id DESC").Find(&machines).Error
	if err != nil {
		return nil, err
	}
	return machines, nil
}

// UpdateStatus 更新服务器机器状态
func (r *serverMachineRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.ServerMachine{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateLoad 更新服务器机器负载
func (r *serverMachineRepository) UpdateLoad(id uint, cpu, memory, disk float64) error {
	return r.db.Model(&model.ServerMachine{}).Where("id = ?", id).Updates(map[string]interface{}{
		"cpu":    cpu,
		"memory": memory,
		"disk":   disk,
	}).Error
}

// GetByToken 根据Token获取服务器机器
func (r *serverMachineRepository) GetByToken(token string) (*model.ServerMachine, error) {
	var machine model.ServerMachine
	err := r.db.Where("token = ?", token).First(&machine).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &machine, nil
}

// UpdateToken 更新服务器机器Token
func (r *serverMachineRepository) UpdateToken(id uint, token string) error {
	return r.db.Model(&model.ServerMachine{}).Where("id = ?", id).Update("token", token).Error
}

// UpdateMachineStatus 更新服务器机器状态（来自节点上报）
func (r *serverMachineRepository) UpdateMachineStatus(id uint, cpu, memory, disk float64, uptime int64) error {
	now := time.Now()
	return r.db.Model(&model.ServerMachine{}).Where("id = ?", id).Updates(map[string]interface{}{
		"cpu":           cpu,
		"memory":        memory,
		"disk":          disk,
		"uptime":        uptime,
		"status":        1,
		"last_check_at": now,
	}).Error
}

// ServerMachineLoadHistoryRepository 服务器机器负载历史仓储接口
type ServerMachineLoadHistoryRepository interface {
	Create(history *model.ServerMachineLoadHistory) error
	List(page, pageSize int, machineID uint) ([]model.ServerMachineLoadHistory, int64, error)
	GetLatest(machineID uint) (*model.ServerMachineLoadHistory, error)
}

type serverMachineLoadHistoryRepository struct {
	db *gorm.DB
}

// NewServerMachineLoadHistoryRepository 创建服务器机器负载历史仓储
func NewServerMachineLoadHistoryRepository() ServerMachineLoadHistoryRepository {
	return &serverMachineLoadHistoryRepository{
		db: database.Get(),
	}
}

// Create 创建负载历史
func (r *serverMachineLoadHistoryRepository) Create(history *model.ServerMachineLoadHistory) error {
	return r.db.Create(history).Error
}

// List 负载历史列表
func (r *serverMachineLoadHistoryRepository) List(page, pageSize int, machineID uint) ([]model.ServerMachineLoadHistory, int64, error) {
	var histories []model.ServerMachineLoadHistory
	var total int64

	query := r.db.Model(&model.ServerMachineLoadHistory{}).Where("machine_id = ?", machineID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// GetLatest 获取最新的负载记录
func (r *serverMachineLoadHistoryRepository) GetLatest(machineID uint) (*model.ServerMachineLoadHistory, error) {
	var history model.ServerMachineLoadHistory
	err := r.db.Where("machine_id = ?", machineID).Order("id DESC").First(&history).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}
