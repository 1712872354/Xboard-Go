package model

import (
	"time"
)

// SubscribeTemplate 订阅模板模型
type SubscribeTemplate struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`        // 模板名称
	Content     string    `gorm:"type:text;not null" json:"content"`             // 模板内容
	Type        string    `gorm:"type:varchar(50)" json:"type"`                  // 类型：clash, v2ray, singbox
	Description string    `gorm:"type:text" json:"description"`                  // 描述
	Sort        int       `gorm:"default:0" json:"sort"`                         // 排序值
	Status      int       `gorm:"default:1" json:"status"`                       // 状态：1启用，0禁用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (SubscribeTemplate) TableName() string {
	return "subscribe_templates"
}

// 订阅模板类型常量
const (
	SubscribeTemplateTypeClash   = "clash"
	SubscribeTemplateTypeV2Ray   = "v2ray"
	SubscribeTemplateTypeSingBox = "singbox"
)

// IsActive 是否激活
func (st *SubscribeTemplate) IsActive() bool {
	return st.Status == 1
}
