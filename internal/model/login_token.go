package model

import (
	"time"
)

// QuickLoginToken 快速登录Token模型
type QuickLoginToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Token     string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (QuickLoginToken) TableName() string {
	return "quick_login_tokens"
}

// MailLoginToken 邮件登录Token模型
type MailLoginToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Token     string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (MailLoginToken) TableName() string {
	return "mail_login_tokens"
}
