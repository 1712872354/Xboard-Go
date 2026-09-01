package model

import "time"

// Plan 套餐模型
type Plan struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	Name               string    `gorm:"type:varchar(100);not null" json:"name"`
	Price              float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Traffic            int64     `gorm:"not null" json:"traffic"`          // 流量（字节）
	DurationDays       int       `gorm:"not null" json:"duration_days"`     // 有效期（天）
	DeviceLimit        int       `gorm:"default:0" json:"device_limit"`    // 设备数限制，0为不限
	NodeGroup          string    `gorm:"type:varchar(100)" json:"node_group"` // 节点组
	ResetTrafficMethod int       `gorm:"default:5" json:"reset_traffic_method"` // 流量重置方式，5=跟随系统
	Description        string    `gorm:"type:text" json:"description"`
	Status             int       `gorm:"default:1" json:"status"`          // 1: 上架, 0: 下架
	SpeedLimit         *int      `json:"speed_limit"`                     // 套餐级限速（Mbps）
	Sell               bool      `gorm:"default:true" json:"sell"`        // 是否允许购买
	Renew              bool      `gorm:"default:true" json:"renew"`       // 是否允许续费
	CapacityLimit      *int      `json:"capacity_limit"`                  // 订阅人数限制
	Tags               string    `gorm:"type:varchar(500)" json:"tags"`   // 套餐标签（逗号分隔）
	Sort               int       `gorm:"default:0" json:"sort"`           // 排序
	Content            string    `gorm:"type:text" json:"content"`        // 套餐说明（Markdown）
	Prices             string    `gorm:"type:text" json:"prices"`         // 多周期价格（JSON）
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Plan) TableName() string {
	return "plans"
}

// IsAvailable 套餐是否可用
func (p *Plan) IsAvailable() bool {
	return p.Status == 1
}
