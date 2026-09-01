package model

import (
	"time"
)

// TicketMessage 工单消息模型
type TicketMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TicketID  uint      `gorm:"index;not null" json:"ticket_id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	IsAdmin   bool      `gorm:"default:false" json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Ticket Ticket `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (TicketMessage) TableName() string {
	return "ticket_messages"
}
