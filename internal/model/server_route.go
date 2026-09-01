package model

import (
	"time"
)

// ServerRoute 服务器路由模型
type ServerRoute struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	GroupID     uint      `gorm:"index;not null" json:"group_id"`         // 关联分组ID
	Name        string    `gorm:"type:varchar(100);not null" json:"name"` // 路由名称（remarks）
	Match       string    `gorm:"type:text;not null" json:"match"`        // 匹配规则（换行分隔文本）
	Action      string    `gorm:"type:varchar(50);not null" json:"action"` // 动作：block, direct, dns, proxy
	ActionValue string    `gorm:"type:varchar(255)" json:"action_value"`  // 动作值（dns时为DNS服务器，proxy时为outbound tag）
	Sort        int       `gorm:"default:0" json:"sort"`                  // 排序值
	Status      int       `gorm:"default:1" json:"status"`                // 状态：1启用，0禁用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ServerRoute) TableName() string {
	return "server_routes"
}

// 路由动作常量
const (
	RouteActionBlock  = "block"  // 阻止访问
	RouteActionDirect = "direct" // 直连
	RouteActionDNS    = "dns"    // DNS解析
	RouteActionProxy  = "proxy"  // 转发
)

// IsActive 是否激活
func (sr *ServerRoute) IsActive() bool {
	return sr.Status == 1
}
