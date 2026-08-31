package model

import "time"

// 订单状态
const (
	OrderStatusPending   = 0 // 待支付
	OrderStatusPaid      = 1 // 已支付
	OrderStatusCancelled = 2 // 已取消
	OrderStatusRefunded  = 3 // 已退款
)

// Order 订单模型
type Order struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `gorm:"index;not null" json:"user_id"`
	PlanID        uint       `gorm:"not null" json:"plan_id"`
	Amount        float64    `gorm:"type:decimal(10,2);not null" json:"amount"`           // 原始金额（分）
	CouponCode    string     `gorm:"type:varchar(50)" json:"coupon_code"`                 // 优惠券码
	Discount      float64    `gorm:"type:decimal(10,2);default:0" json:"discount"`        // 折扣金额（分）
	ActualAmount  float64    `gorm:"type:decimal(10,2)" json:"actual_amount"`             // 实付金额（分）
	Status        int        `gorm:"default:0;index" json:"status"`                       // 0: pending, 1: paid, 2: cancelled, 3: refunded
	PaymentMethod string     `gorm:"type:varchar(50)" json:"payment_method"`
	TradeNo       string     `gorm:"type:varchar(100);uniqueIndex" json:"trade_no"`       // 商户订单号
	Commission    float64    `gorm:"type:decimal(10,2);default:0" json:"commission"`      // 佣金金额
	CreatedAt     time.Time  `json:"created_at"`
	PaidAt        *time.Time `json:"paid_at"`

	Plan Plan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// IsPaid 是否已支付
func (o *Order) IsPaid() bool {
	return o.Status == OrderStatusPaid
}

// IsPending 是否待支付
func (o *Order) IsPending() bool {
	return o.Status == OrderStatusPending
}
