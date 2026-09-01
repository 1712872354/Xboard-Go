package repository

import (
	"errors"
	"time"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// TicketRepository 工单仓储接口
type TicketRepository interface {
	// Ticket 操作
	Create(ticket *model.Ticket) error
	GetByID(id uint) (*model.Ticket, error)
	Update(ticket *model.Ticket) error
	Delete(id uint) error

	// 列表查询
	ListByUser(userID uint, page, pageSize int, status int) ([]model.Ticket, int64, error)
	ListAll(page, pageSize int, status int, category int, keyword string) ([]model.Ticket, int64, error)

	// 回复操作
	AddReply(reply *model.TicketReply) error
	GetReplies(ticketID uint) ([]model.TicketReply, error)

	// 统计
	CountByStatus(status int) (int64, error)
	CountByUserAndStatus(userID uint, status int) (int64, error)
}

type ticketRepository struct {
	db *gorm.DB
}

// NewTicketRepository 创建工单仓储
func NewTicketRepository() TicketRepository {
	return &ticketRepository{
		db: database.Get(),
	}
}

// Create 创建工单
func (r *ticketRepository) Create(ticket *model.Ticket) error {
	return r.db.Create(ticket).Error
}

// GetByID 根据ID获取工单
func (r *ticketRepository) GetByID(id uint) (*model.Ticket, error) {
	var ticket model.Ticket
	err := r.db.Preload("User").First(&ticket, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ticket, nil
}

// Update 更新工单
func (r *ticketRepository) Update(ticket *model.Ticket) error {
	return r.db.Save(ticket).Error
}

// Delete 删除工单
func (r *ticketRepository) Delete(id uint) error {
	return r.db.Delete(&model.Ticket{}, id).Error
}

// ListByUser 用户工单列表
func (r *ticketRepository) ListByUser(userID uint, page, pageSize int, status int) ([]model.Ticket, int64, error) {
	var tickets []model.Ticket
	var total int64

	query := r.db.Model(&model.Ticket{}).Where("user_id = ?", userID)

	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// ListAll 所有工单列表（管理员）
func (r *ticketRepository) ListAll(page, pageSize int, status int, category int, keyword string) ([]model.Ticket, int64, error) {
	var tickets []model.Ticket
	var total int64

	query := r.db.Model(&model.Ticket{}).Preload("User")

	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if category >= 0 {
		query = query.Where("category = ?", category)
	}
	if keyword != "" {
		query = query.Joins("LEFT JOIN users ON users.id = tickets.user_id").
			Where("users.email LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("status ASC, priority DESC, updated_at DESC").
		Offset(offset).Limit(pageSize).Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// AddReply 添加回复
func (r *ticketRepository) AddReply(reply *model.TicketReply) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 创建回复
		if err := tx.Create(reply).Error; err != nil {
			return err
		}

		// 更新工单最后回复时间和状态
		now := time.Now()
		updates := map[string]interface{}{
			"last_reply": now,
			"updated_at": now,
		}

		// 如果是管理员回复，状态变为"已回复"
		if reply.IsAdmin {
			updates["status"] = model.TicketStatusReplied
		} else {
			// 用户回复，状态变为"待处理"
			updates["status"] = model.TicketStatusOpen
		}

		if err := tx.Model(&model.Ticket{}).Where("id = ?", reply.TicketID).Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetReplies 获取工单回复列表
func (r *ticketRepository) GetReplies(ticketID uint) ([]model.TicketReply, error) {
	var replies []model.TicketReply
	err := r.db.Where("ticket_id = ?", ticketID).
		Preload("User").
		Order("created_at ASC").
		Find(&replies).Error
	return replies, err
}

// CountByStatus 按状态统计工单数
func (r *ticketRepository) CountByStatus(status int) (int64, error) {
	var count int64
	query := r.db.Model(&model.Ticket{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	err := query.Count(&count).Error
	return count, err
}

// CountByUserAndStatus 按用户和状态统计
func (r *ticketRepository) CountByUserAndStatus(userID uint, status int) (int64, error) {
	var count int64
	query := r.db.Model(&model.Ticket{}).Where("user_id = ?", userID)
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	err := query.Count(&count).Error
	return count, err
}
