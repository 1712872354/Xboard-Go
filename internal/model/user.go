package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Email           string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash    string     `gorm:"type:varchar(255);not null" json:"-"`
	Role            string     `gorm:"type:varchar(20);default:'user'" json:"role"` // user, admin
	Status          int        `gorm:"default:1" json:"status"`                     // 1: active, 0: banned
	TrafficLimit    int64      `gorm:"default:0" json:"traffic_limit"`              // 流量限制（字节）
	UsedTraffic     int64      `gorm:"default:0" json:"used_traffic"`               // 已用流量（字节）
	ExpiredAt       *time.Time `json:"expired_at"`                                  // 套餐到期时间
	SubscribeToken  string     `gorm:"type:varchar(64);uniqueIndex" json:"-"`       // 订阅token
	TwoFactorSecret string     `gorm:"type:varchar(100)" json:"-"`                  // 2FA密钥
	TwoFactorEnabled bool      `gorm:"default:false" json:"two_factor_enabled"`     // 是否启用2FA
	InviteCodeID    *uint      `json:"invite_code_id"`                              // 使用的邀请码ID
	Balance         float64    `gorm:"default:0" json:"balance"`                    // 账户余额
	Commission      float64    `gorm:"default:0" json:"commission"`                 // 佣金余额
	TelegramID      int64      `gorm:"default:0" json:"telegram_id"`                // Telegram 用户 ID
	LastResetAt     *time.Time `json:"last_reset_at"`                               // 上次流量重置时间
	NextResetAt     *time.Time `json:"next_reset_at"`                               // 下次流量重置时间
	ResetCount      int        `gorm:"default:0" json:"reset_count"`                // 流量重置次数
	RemindExpire    bool       `gorm:"default:true" json:"remind_expire"`           // 到期提醒
	RemindTraffic   bool       `gorm:"default:true" json:"remind_traffic"`          // 流量提醒
	OnlineCount     int        `gorm:"default:0" json:"online_count"`               // 在线设备数
	LastOnlineAt    *time.Time `json:"last_online_at"`                              // 最后在线时间
	PlanID          *uint      `json:"plan_id"`                                     // 当前套餐ID
	UUID            string     `gorm:"type:varchar(36);uniqueIndex" json:"uuid"`      // 用户UUID
	InviterID       *uint      `gorm:"index" json:"inviter_id"`                       // 邀请人用户ID
	CommissionType  int        `gorm:"default:0" json:"commission_type"`               // 佣金类型：0=系统 1=按周期 2=一次性
	CommissionRate  *int       `json:"commission_rate"`                                // 返佣比例（0-100）
	Discount        *int       `json:"discount"`                                       // 专享折扣比例
	SpeedLimit      *int       `json:"speed_limit"`                                    // 用户级限速（Mbps）
	DeviceLimit     *int       `json:"device_limit"`                                   // 用户级设备限制
	GroupID         *uint      `gorm:"index" json:"group_id"`                          // 权限组ID
	Remarks         string     `gorm:"type:text" json:"remarks"`                       // 管理员备注
	LastLoginAt     *time.Time `json:"last_login_at"`                                  // 最后登录时间
	LastLoginIP     string     `gorm:"type:varchar(45)" json:"last_login_ip"`          // 最后登录IP
	IsStaff         bool       `gorm:"default:false" json:"is_staff"`                  // 员工标识
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsActive 用户是否活跃
func (u *User) IsActive() bool {
	return u.Status == 1
}

// IsAdmin 是否管理员
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// HasExpired 套餐是否过期
func (u *User) HasExpired() bool {
	if u.ExpiredAt == nil {
		return true
	}
	return time.Now().After(*u.ExpiredAt)
}

// HasEnoughTraffic 是否有足够流量
func (u *User) HasEnoughTraffic() bool {
	if u.TrafficLimit <= 0 {
		return true // 0 表示不限流量
	}
	return u.UsedTraffic < u.TrafficLimit
}

// CanUseService 是否可以使用服务
func (u *User) CanUseService() bool {
	return u.IsActive() && !u.HasExpired() && u.HasEnoughTraffic()
}

// ShouldResetTraffic 检查用户是否需要重置流量
func (u *User) ShouldResetTraffic() bool {
	if u.NextResetAt == nil {
		return false
	}
	return time.Now().After(*u.NextResetAt) || time.Now().Equal(*u.NextResetAt)
}

// TrafficResetMethod 流量重置方式常量
const (
	ResetTrafficNever         = 0 // 永不重置
	ResetTrafficFirstDayMonth = 1 // 每月1号
	ResetTrafficMonthly       = 2 // 按月（按开通日）
	ResetTrafficFirstDayYear  = 3 // 每年1月1号
	ResetTrafficYearly        = 4 // 按年（按开通日）
	ResetTrafficFollowSystem  = 5 // 跟随系统设置
)
