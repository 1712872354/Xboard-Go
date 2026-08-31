package model

import (
	"time"
)

// GiftCardTemplate 礼品卡模板
type GiftCardTemplate struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`     // 模板名称
	Description string    `gorm:"type:text" json:"description"`               // 描述
	Type        int       `gorm:"default:1" json:"type"`                      // 类型：1金额，2流量，3时长
	Value       float64   `gorm:"default:0" json:"value"`                     // 值
	Traffic     int64     `gorm:"default:0" json:"traffic"`                   // 流量（字节）
	Duration    int       `gorm:"default:0" json:"duration"`                  // 时长（天）
	PlanID      *uint     `json:"plan_id"`                                    // 关联套餐ID
	Price       float64   `gorm:"default:0" json:"price"`                     // 售价
	Status      int       `gorm:"default:1" json:"status"`                    // 状态：1启用，0禁用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (GiftCardTemplate) TableName() string {
	return "gift_card_templates"
}

// GiftCardCode 礼品卡码
type GiftCardCode struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	TemplateID uint       `gorm:"index;not null" json:"template_id"` // 关联模板ID
	Code       string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"` // 礼品码
	Status     int        `gorm:"default:0" json:"status"`           // 状态：0未使用，1已使用
	UserID     *uint      `gorm:"index" json:"user_id"`              // 使用者ID
	UsedAt     *time.Time `json:"used_at"`                           // 使用时间
	OrderID    *uint      `json:"order_id"`                          // 关联订单ID
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (GiftCardCode) TableName() string {
	return "gift_card_codes"
}

// IsUsed 是否已使用
func (g *GiftCardCode) IsUsed() bool {
	return g.Status == 1
}

// GiftCardUsage 礼品卡使用记录
type GiftCardUsage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CodeID    uint      `gorm:"index;not null" json:"code_id"`     // 礼品码ID
	UserID    uint      `gorm:"index;not null" json:"user_id"`     // 用户ID
	OrderID   uint      `json:"order_id"`                          // 订单ID
	Amount    float64   `gorm:"default:0" json:"amount"`            // 金额
	Traffic   int64     `gorm:"default:0" json:"traffic"`           // 流量
	Duration  int       `gorm:"default:0" json:"duration"`          // 时长
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (GiftCardUsage) TableName() string {
	return "gift_card_usages"
}

// 礼品卡类型常量
const (
	GiftCardTypeAmount   = 1 // 金额
	GiftCardTypeTraffic  = 2 // 流量
	GiftCardTypeDuration = 3 // 时长
)
