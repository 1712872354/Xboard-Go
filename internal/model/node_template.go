package model

import "time"

// NodeTemplate 节点模板模型
type NodeTemplate struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Type        string    `gorm:"type:varchar(20);not null" json:"type"`
	ServerInfo  string    `gorm:"type:text" json:"server_info"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (NodeTemplate) TableName() string {
	return "node_templates"
}
