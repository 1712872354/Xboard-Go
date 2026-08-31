package model

import (
	"time"
)

// InviteCode 邀请码模型
type InviteCode struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`      // 邀请人ID
	Code        string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"` // 邀请码
	Commission  float64   `gorm:"default:0" json:"commission"`         // 佣金比例（百分比）
	UsedCount   int       `gorm:"default:0" json:"used_count"`         // 已使用次数
	LimitCount  int       `gorm:"default:0" json:"limit_count"`        // 限制使用次数，0表示不限制
	Status      int       `gorm:"default:1" json:"status"`             // 状态：1启用，0禁用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (InviteCode) TableName() string {
	return "invite_codes"
}

// IsActive 是否激活
func (i *InviteCode) IsActive() bool {
	if i.Status != 1 {
		return false
	}
	if i.LimitCount > 0 && i.UsedCount >= i.LimitCount {
		return false
	}
	return true
}

// CommissionLog 佣金记录
type CommissionLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`     // 获得佣金的用户ID
	FromUserID uint    `gorm:"index;not null" json:"from_user_id"` // 来源用户ID
	OrderID   uint      `gorm:"index;not null" json:"order_id"`    // 关联订单ID
	Amount    float64   `gorm:"default:0" json:"amount"`            // 佣金金额
	OrderAmount float64 `gorm:"default:0" json:"order_amount"`      // 订单金额
	Commission float64  `gorm:"default:0" json:"commission"`        // 佣金比例
	Status    int       `gorm:"default:0" json:"status"`            // 状态：0待结算，1已结算，2已取消
	Remark    string    `gorm:"type:varchar(255)" json:"remark"`    // 备注
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (CommissionLog) TableName() string {
	return "commission_logs"
}

// 佣金状态常量
const (
	CommissionStatusPending  = 0 // 待结算
	CommissionStatusSettled  = 1 // 已结算
	CommissionStatusCanceled = 2 // 已取消
)
