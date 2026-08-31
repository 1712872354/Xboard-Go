package model

import (
	"time"
)

// MailTemplate 邮件模板模型
type MailTemplate struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"` // 模板名称（唯一标识）
	Subject   string    `gorm:"type:varchar(255);not null" json:"subject"`          // 邮件主题
	Body      string    `gorm:"type:text;not null" json:"body"`                     // 邮件内容（支持HTML）
	Language  string    `gorm:"type:varchar(20);default:'zh-CN'" json:"language"`   // 语言
	Remark    string    `gorm:"type:varchar(255)" json:"remark"`                    // 备注说明
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (MailTemplate) TableName() string {
	return "mail_templates"
}

// 邮件模板名称常量
const (
	MailTemplateRegister        = "register"         // 注册验证
	MailTemplateResetPassword   = "reset_password"   // 重置密码
	MailTemplateOrderCreated    = "order_created"    // 订单创建
	MailTemplateOrderPaid       = "order_paid"       // 订单支付成功
	MailTemplateTrafficWarning  = "traffic_warning"  // 流量警告
	MailTemplateAccountExpired  = "account_expired"  // 账号到期提醒
	MailTemplateWelcome         = "welcome"          // 欢迎邮件
	MailTemplateTicketReply     = "ticket_reply"     // 工单回复
	MailTemplateCustom          = "custom"           // 自定义邮件
)
