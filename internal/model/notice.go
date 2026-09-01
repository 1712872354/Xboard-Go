package model

import (
	"time"
)

// Notice 公告模型
type Notice struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	ImgURL    string    `gorm:"type:varchar(500)" json:"img_url"`                // 公告图片
	Tags      string    `gorm:"type:varchar(500)" json:"tags"`                   // 标签，逗号分隔
	Show      int       `gorm:"default:1" json:"show"`                           // 是否显示：1显示，0隐藏
	Sort      int       `gorm:"default:0" json:"sort"`                           // 排序值
	Groups    string    `gorm:"type:varchar(255)" json:"groups"`                 // 可见用户组，空表示所有
	Popup     bool      `gorm:"default:false" json:"popup"`                    // 是否弹窗显示
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Notice) TableName() string {
	return "notices"
}

// IsVisible 是否可见
func (n *Notice) IsVisible() bool {
	return n.Show == 1
}
