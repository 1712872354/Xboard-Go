package model

import (
	"time"
)

// MailLog 邮件日志模型
type MailLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	To        string    `gorm:"type:varchar(255);not null" json:"to"`         // 收件人
	Subject   string    `gorm:"type:varchar(255);not null" json:"subject"`    // 邮件主题
	Body      string    `gorm:"type:text" json:"body"`                        // 邮件内容
	Status    int       `gorm:"default:0" json:"status"`                      // 状态：0待发送，1已发送，2发送失败
	Error     string    `gorm:"type:text" json:"error"`                       // 错误信息
	UserID    *uint     `gorm:"index" json:"user_id"`                         // 关联用户ID
	Template  string    `gorm:"type:varchar(100)" json:"template"`            // 使用的模板
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (MailLog) TableName() string {
	return "mail_logs"
}

// 邮件状态常量
const (
	MailStatusPending = 0 // 待发送
	MailStatusSent    = 1 // 已发送
	MailStatusFailed  = 2 // 发送失败
)

// IsSent 是否已发送
func (m *MailLog) IsSent() bool {
	return m.Status == MailStatusSent
}

// IsFailed 是否发送失败
func (m *MailLog) IsFailed() bool {
	return m.Status == MailStatusFailed
}

// IsPending 是否待发送
func (m *MailLog) IsPending() bool {
	return m.Status == MailStatusPending
}
