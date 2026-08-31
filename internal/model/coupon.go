package model

import (
	"time"
)

// Coupon 优惠券模型
type Coupon struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Code         string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"` // 优惠码
	Name         string    `gorm:"type:varchar(100);not null" json:"name"`            // 优惠券名称
	Type         int       `gorm:"default:1" json:"type"`                             // 类型：1金额，2百分比
	Value        float64   `gorm:"default:0" json:"value"`                            // 值：金额或百分比
	MinAmount    float64   `gorm:"default:0" json:"min_amount"`                       // 最低消费金额
	MaxDiscount  float64   `gorm:"default:0" json:"max_discount"`                     // 最大优惠金额（百分比类型时生效）
	PlanIDs      string    `gorm:"type:text" json:"plan_ids"`                         // 适用套餐ID，空表示所有，逗号分隔
	UserIDs      string    `gorm:"type:text" json:"user_ids"`                         // 适用用户ID，空表示所有，逗号分隔
	UsedCount    int       `gorm:"default:0" json:"used_count"`                       // 已使用次数
	LimitCount   int       `gorm:"default:0" json:"limit_count"`                      // 限制使用次数，0表示不限制
	StartDate    *time.Time `json:"start_date"`                                       // 开始时间
	EndDate      *time.Time `json:"end_date"`                                         // 结束时间
	Status       int       `gorm:"default:1" json:"status"`                           // 状态：1启用，0禁用
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Coupon) TableName() string {
	return "coupons"
}

// 优惠券类型常量
const (
	CouponTypeAmount   = 1 // 金额优惠
	CouponTypePercent  = 2 // 百分比优惠
)

// IsActive 是否激活
func (c *Coupon) IsActive() bool {
	if c.Status != 1 {
		return false
	}

	now := time.Now()

	if c.StartDate != nil && now.Before(*c.StartDate) {
		return false
	}

	if c.EndDate != nil && now.After(*c.EndDate) {
		return false
	}

	if c.LimitCount > 0 && c.UsedCount >= c.LimitCount {
		return false
	}

	return true
}

// CalculateDiscount 计算优惠金额
func (c *Coupon) CalculateDiscount(amount float64) float64 {
	if !c.IsActive() {
		return 0
	}

	if amount < c.MinAmount {
		return 0
	}

	var discount float64

	switch c.Type {
	case CouponTypeAmount:
		discount = c.Value
	case CouponTypePercent:
		discount = amount * c.Value / 100
		if c.MaxDiscount > 0 && discount > c.MaxDiscount {
			discount = c.MaxDiscount
		}
	}

	if discount > amount {
		discount = amount
	}

	return discount
}
