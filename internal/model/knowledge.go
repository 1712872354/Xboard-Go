package model

import (
	"time"
)

// Knowledge 知识库模型
type Knowledge struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Category  string    `gorm:"type:varchar(100);index;not null" json:"category"` // 分类
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Show      int       `gorm:"default:1" json:"show"`       // 是否显示：1显示，0隐藏
	Sort      int       `gorm:"default:0" json:"sort"`       // 排序值
	Language  string    `gorm:"type:varchar(20);default:'zh-CN'" json:"language"` // 语言
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Knowledge) TableName() string {
	return "knowledges"
}

// IsVisible 是否可见
func (k *Knowledge) IsVisible() bool {
	return k.Show == 1
}
