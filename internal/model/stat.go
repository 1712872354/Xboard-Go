package model

import "time"

// Stat 每日统计模型
type Stat struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	RecordAt         time.Time `gorm:"index;not null" json:"record_at"`          // 统计日期
	RecordType       string    `gorm:"type:varchar(10);default:'d'" json:"record_type"` // d=day
	OrderCount       int64     `gorm:"default:0" json:"order_count"`             // 订单总数
	PaidCount        int64     `gorm:"default:0" json:"paid_count"`              // 已支付订单数
	OrderTotal       float64   `gorm:"default:0" json:"order_total"`             // 订单总金额（分）
	PaidTotal        float64   `gorm:"default:0" json:"paid_total"`              // 已支付金额（分）
	RegisterCount    int64     `gorm:"default:0" json:"register_count"`          // 注册数
	InviteCount      int64     `gorm:"default:0" json:"invite_count"`            // 邀请注册数
	CommissionCount  int64     `gorm:"default:0" json:"commission_count"`        // 佣金笔数
	CommissionTotal  float64   `gorm:"default:0" json:"commission_total"`        // 佣金总额（分）
	TransferUsed     int64     `gorm:"default:0" json:"transfer_used"`           // 流量使用总量（字节）
	CreatedAt        time.Time `json:"created_at"`
}

// TableName 指定表名
func (Stat) TableName() string {
	return "stats"
}
