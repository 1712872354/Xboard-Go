package model

import (
	"time"
)

// AdminAuditLog 管理员审计日志模型
type AdminAuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`      // 操作者ID
	Username  string    `gorm:"type:varchar(100)" json:"username"`   // 操作者用户名
	Action    string    `gorm:"type:varchar(100);not null" json:"action"` // 操作类型
	Resource  string    `gorm:"type:varchar(100)" json:"resource"`   // 资源类型
	ResourceID uint     `json:"resource_id"`                         // 资源ID
	Detail    string    `gorm:"type:text" json:"detail"`             // 操作详情
	IP        string    `gorm:"type:varchar(50)" json:"ip"`          // IP地址
	UserAgent string    `gorm:"type:varchar(500)" json:"user_agent"` // UserAgent
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (AdminAuditLog) TableName() string {
	return "admin_audit_logs"
}

// 审计日志操作类型常量
const (
	AuditActionCreate = "create" // 创建
	AuditActionUpdate = "update" // 更新
	AuditActionDelete = "delete" // 删除
	AuditActionLogin  = "login"  // 登录
	AuditActionLogout = "logout" // 登出
	AuditActionExport = "export" // 导出
	AuditActionImport = "import" // 导入
)
