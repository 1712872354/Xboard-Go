package model

import (
	"time"
)

// ServerGroup 服务器分组模型
type ServerGroup struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`        // 分组名称
	Description string    `gorm:"type:text" json:"description"`                  // 描述
	PlanIDs     string    `gorm:"type:text" json:"plan_ids"`                     // 关联套餐ID，空表示所有，逗号分隔
	Sort        int       `gorm:"default:0" json:"sort"`                         // 排序值
	Status      int       `gorm:"default:1" json:"status"`                       // 状态：1启用，0禁用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ServerGroup) TableName() string {
	return "server_groups"
}

// IsActive 是否激活
func (sg *ServerGroup) IsActive() bool {
	return sg.Status == 1
}
