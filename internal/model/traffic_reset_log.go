package model

import "time"

// TrafficResetLog 流量重置日志
type TrafficResetLog struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `gorm:"index;not null" json:"user_id"`
	ResetType     string     `gorm:"type:varchar(20);not null" json:"reset_type"`      // first_day_month, monthly, first_day_year, yearly, manual
	TriggerSource string     `gorm:"type:varchar(20);not null" json:"trigger_source"`  // auto, manual, cron
	OldUpload     int64      `gorm:"default:0" json:"old_upload"`
	OldDownload   int64      `gorm:"default:0" json:"old_download"`
	OldTotal      int64      `gorm:"default:0" json:"old_total"`
	NewUpload     int64      `gorm:"default:0" json:"new_upload"`
	NewDownload   int64      `gorm:"default:0" json:"new_download"`
	NewTotal      int64      `gorm:"default:0" json:"new_total"`
	Metadata      string     `gorm:"type:text" json:"metadata"` // JSON 格式的额外信息
	ResetTime     time.Time  `gorm:"not null" json:"reset_time"`
	CreatedAt     time.Time  `json:"created_at"`
}

// TableName 指定表名
func (TrafficResetLog) TableName() string {
	return "traffic_reset_logs"
}

// 流量重置触发源常量
const (
	TriggerSourceAuto   = "auto"
	TriggerSourceManual = "manual"
	TriggerSourceCron   = "cron"
)

// 流量重置类型常量
const (
	ResetTypeFirstDayMonth = "first_day_month"
	ResetTypeMonthly       = "monthly"
	ResetTypeFirstDayYear  = "first_day_year"
	ResetTypeYearly        = "yearly"
	ResetTypeManual        = "manual"
)
