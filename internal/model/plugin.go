package model

import (
	"time"
)

// Plugin 插件模型
type Plugin struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"` // 插件名称（唯一标识）
	Title       string    `gorm:"type:varchar(100);not null" json:"title"`            // 显示名称
	Description string    `gorm:"type:text" json:"description"`                       // 描述
	Version     string    `gorm:"type:varchar(20)" json:"version"`                    // 版本号
	Author      string    `gorm:"type:varchar(100)" json:"author"`                    // 作者
	Homepage    string    `gorm:"type:varchar(255)" json:"homepage"`                  // 主页
	Config      string    `gorm:"type:text" json:"config"`                            // 配置（JSON格式）
	Type        string     `gorm:"type:varchar(20);default:'feature'" json:"type"`    // 插件类型：feature/payment
	IsProtected bool       `gorm:"default:false" json:"is_protected"`                 // 是否受保护（系统内置插件不可删除）
	Status      int        `gorm:"default:0" json:"status"`                           // 状态：0禁用，1启用
	InstalledAt *time.Time `json:"installed_at"`                                      // 安装时间
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Plugin) TableName() string {
	return "plugins"
}

// IsEnabled 是否启用
func (p *Plugin) IsEnabled() bool {
	return p.Status == 1
}
