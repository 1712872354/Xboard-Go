package model

import "time"

// RedeemCode 兑换码/卡密模型
type RedeemCode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	PlanID    uint      `gorm:"not null" json:"plan_id"`
	Status    int       `gorm:"default:0" json:"status"` // 0: 未使用, 1: 已使用
	UsedBy    *uint     `json:"used_by"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time `json:"created_at"`

	Plan Plan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

// TableName 指定表名
func (RedeemCode) TableName() string {
	return "redeem_codes"
}

// IsUsed 是否已使用
func (r *RedeemCode) IsUsed() bool {
	return r.Status == 1
}
