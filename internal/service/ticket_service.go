package service

import (
	"errors"
	"fmt"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// TicketService 工单服务接口
type TicketService interface {
	// CreateTicket 创建工单
	CreateTicket(userID uint, subject, content string, category, priority int) (*model.Ticket, error)
	// GetTicket 获取工单详情
	GetTicket(id uint, userID uint, isAdmin bool) (*model.Ticket, []model.TicketReply, error)
	// Reply 回复工单
	Reply(ticketID uint, userID uint, content string, isAdmin bool) error
	// CloseTicket 关闭工单
	CloseTicket(ticketID uint, userID uint, isAdmin bool) error
	// ListUserTickets 用户工单列表
	ListUserTickets(userID uint, page, pageSize int, status int) ([]model.Ticket, int64, error)
	// ListAllTickets 所有工单列表（管理员）
	ListAllTickets(page, pageSize int, status int, category int) ([]model.Ticket, int64, error)
	// GetStats 获取工单统计
	GetStats(userID uint, isAdmin bool) (*TicketStats, error)
	// DeleteTicket 删除工单（管理员）
	DeleteTicket(ticketID uint) error
}

type ticketService struct {
	ticketRepo repository.TicketRepository
	userRepo   repository.UserRepository
}

// TicketStats 工单统计
type TicketStats struct {
	Open    int64 `json:"open"`    // 待处理
	Replied int64 `json:"replied"` // 已回复
	Closed  int64 `json:"closed"`  // 已关闭
	Total   int64 `json:"total"`   // 总数
}

// NewTicketService 创建工单服务
func NewTicketService(ticketRepo repository.TicketRepository, userRepo repository.UserRepository) TicketService {
	return &ticketService{
		ticketRepo: ticketRepo,
		userRepo:   userRepo,
	}
}

// CreateTicket 创建工单
func (s *ticketService) CreateTicket(userID uint, subject, content string, category, priority int) (*model.Ticket, error) {
	if subject == "" {
		return nil, errors.New("subject is required")
	}
	if content == "" {
		return nil, errors.New("content is required")
	}
	if len(subject) > 200 {
		return nil, errors.New("subject too long (max 200 chars)")
	}

	// 验证用户存在
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 校验分类和优先级
	if category < model.TicketCategoryGeneral || category > model.TicketCategoryAccount {
		category = model.TicketCategoryGeneral
	}
	if priority < model.TicketPriorityLow || priority > model.TicketPriorityHigh {
		priority = model.TicketPriorityNormal
	}

	ticket := &model.Ticket{
		UserID:   userID,
		Subject:  subject,
		Category: category,
		Priority: priority,
		Status:   model.TicketStatusOpen,
	}

	if err := s.ticketRepo.Create(ticket); err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	// 添加第一条回复（工单内容）
	reply := &model.TicketReply{
		TicketID: ticket.ID,
		UserID:   userID,
		Content:  content,
		IsAdmin:  false,
	}
	if err := s.ticketRepo.AddReply(reply); err != nil {
		return nil, fmt.Errorf("failed to add initial reply: %w", err)
	}

	return ticket, nil
}

// GetTicket 获取工单详情
func (s *ticketService) GetTicket(id uint, userID uint, isAdmin bool) (*model.Ticket, []model.TicketReply, error) {
	ticket, err := s.ticketRepo.GetByID(id)
	if err != nil {
		return nil, nil, err
	}
	if ticket == nil {
		return nil, nil, errors.New("ticket not found")
	}

	// 非管理员只能查看自己的工单
	if !isAdmin && ticket.UserID != userID {
		return nil, nil, errors.New("ticket not found")
	}

	replies, err := s.ticketRepo.GetReplies(id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get replies: %w", err)
	}

	return ticket, replies, nil
}

// Reply 回复工单
func (s *ticketService) Reply(ticketID uint, userID uint, content string, isAdmin bool) error {
	if content == "" {
		return errors.New("content is required")
	}

	ticket, err := s.ticketRepo.GetByID(ticketID)
	if err != nil {
		return err
	}
	if ticket == nil {
		return errors.New("ticket not found")
	}

	// 权限检查
	if !isAdmin && ticket.UserID != userID {
		return errors.New("ticket not found")
	}

	// 状态检查
	if ticket.IsClosed() {
		return errors.New("ticket is closed")
	}

	reply := &model.TicketReply{
		TicketID: ticketID,
		UserID:   userID,
		Content:  content,
		IsAdmin:  isAdmin,
	}

	return s.ticketRepo.AddReply(reply)
}

// CloseTicket 关闭工单
func (s *ticketService) CloseTicket(ticketID uint, userID uint, isAdmin bool) error {
	ticket, err := s.ticketRepo.GetByID(ticketID)
	if err != nil {
		return err
	}
	if ticket == nil {
		return errors.New("ticket not found")
	}

	// 权限检查
	if !isAdmin && ticket.UserID != userID {
		return errors.New("ticket not found")
	}

	if ticket.IsClosed() {
		return errors.New("ticket already closed")
	}

	ticket.Status = model.TicketStatusClosed
	return s.ticketRepo.Update(ticket)
}

// ListUserTickets 用户工单列表
func (s *ticketService) ListUserTickets(userID uint, page, pageSize int, status int) ([]model.Ticket, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.ticketRepo.ListByUser(userID, page, pageSize, status)
}

// ListAllTickets 所有工单列表（管理员）
func (s *ticketService) ListAllTickets(page, pageSize int, status int, category int) ([]model.Ticket, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.ticketRepo.ListAll(page, pageSize, status, category)
}

// GetStats 获取工单统计
func (s *ticketService) GetStats(userID uint, isAdmin bool) (*TicketStats, error) {
	var open, replied, closed int64
	var err error

	if isAdmin {
		open, err = s.ticketRepo.CountByStatus(model.TicketStatusOpen)
		if err != nil {
			return nil, err
		}
		replied, err = s.ticketRepo.CountByStatus(model.TicketStatusReplied)
		if err != nil {
			return nil, err
		}
		closed, err = s.ticketRepo.CountByStatus(model.TicketStatusClosed)
		if err != nil {
			return nil, err
		}
	} else {
		open, err = s.ticketRepo.CountByUserAndStatus(userID, model.TicketStatusOpen)
		if err != nil {
			return nil, err
		}
		replied, err = s.ticketRepo.CountByUserAndStatus(userID, model.TicketStatusReplied)
		if err != nil {
			return nil, err
		}
		closed, err = s.ticketRepo.CountByUserAndStatus(userID, model.TicketStatusClosed)
		if err != nil {
			return nil, err
		}
	}

	return &TicketStats{
		Open:    open,
		Replied: replied,
		Closed:  closed,
		Total:   open + replied + closed,
	}, nil
}

// DeleteTicket 删除工单（管理员）
func (s *ticketService) DeleteTicket(ticketID uint) error {
	ticket, err := s.ticketRepo.GetByID(ticketID)
	if err != nil {
		return err
	}
	if ticket == nil {
		return errors.New("ticket not found")
	}
	return s.ticketRepo.Delete(ticketID)
}
