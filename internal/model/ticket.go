package model

import (
	"time"

	"gorm.io/gorm"
)

// TicketStatus 工单状态
const (
	TicketStatusOpen     = 0 // 待处理
	TicketStatusReplied  = 1 // 已回复（管理员回复后）
	TicketStatusClosed   = 2 // 已关闭
	TicketStatusArchived = 3 // 已归档
)

// TicketPriority 工单优先级
const (
	TicketPriorityLow    = 0 // 低
	TicketPriorityNormal = 1 // 普通
	TicketPriorityHigh   = 2 // 高
)

// TicketCategory 工单分类
const (
	TicketCategoryGeneral   = 0 // 一般问题
	TicketCategoryBilling   = 1 // 账单/支付
	TicketCategoryTechnical = 2 // 技术问题
	TicketCategoryAccount   = 3 // 账户问题
)

// Ticket 工单模型
type Ticket struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	Subject   string         `gorm:"type:varchar(200);not null" json:"subject"`
	Category  int            `gorm:"type:smallint;default:0;index" json:"category"` // 0:一般 1:账单 2:技术 3:账户
	Priority  int            `gorm:"type:smallint;default:1;index" json:"priority"` // 0:低 1:普通 2:高
	Status    int            `gorm:"type:smallint;default:0;index" json:"status"`   // 0:待处理 1:已回复 2:已关闭
	LastReply *time.Time     `gorm:"index" json:"last_reply"`
	Replies   []TicketReply  `gorm:"foreignKey:TicketID" json:"replies,omitempty"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Ticket) TableName() string {
	return "tickets"
}

// IsOpen 是否待处理
func (t *Ticket) IsOpen() bool {
	return t.Status == TicketStatusOpen
}

// IsReplied 是否已回复
func (t *Ticket) IsReplied() bool {
	return t.Status == TicketStatusReplied
}

// IsClosed 是否已关闭
func (t *Ticket) IsClosed() bool {
	return t.Status == TicketStatusClosed
}

// CanReply 是否可以回复
func (t *Ticket) CanReply() bool {
	return t.Status != TicketStatusClosed
}

// TicketReply 工单回复模型
type TicketReply struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	TicketID  uint           `gorm:"index;not null" json:"ticket_id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	IsAdmin   bool           `gorm:"default:false;index" json:"is_admin"` // 是否管理员回复
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (TicketReply) TableName() string {
	return "ticket_replies"
}
