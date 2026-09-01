package model

import (
	"time"
)

// UserToken 用户Token模型（用于会话管理）
type UserToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Token     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"token"`
	IP        string    `gorm:"type:varchar(45)" json:"ip"`
	UserAgent string    `gorm:"type:varchar(500)" json:"user_agent"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (UserToken) TableName() string {
	return "user_tokens"
}

// IsExpired 是否已过期
func (t *UserToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
