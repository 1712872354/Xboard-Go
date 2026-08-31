package model

import "time"

// NodeUser 用户与节点关联
type NodeUser struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index:idx_user_node,unique;not null" json:"user_id"`
	NodeID    uint      `gorm:"index:idx_user_node,unique;not null" json:"node_id"`
	UUID      string    `gorm:"type:varchar(64)" json:"uuid"`
	Passwd    string    `gorm:"type:varchar(100)" json:"passwd"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
	Node Node `gorm:"foreignKey:NodeID" json:"-"`
}

// TableName 指定表名
func (NodeUser) TableName() string {
	return "node_users"
}
