package model

import (
	"time"
)

// Setting 系统设置模型
type Setting struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	Group     string    `gorm:"type:varchar(50);index;default:'general'" json:"group"` // 设置分组：general, email, payment, security, etc.
	Remark    string    `gorm:"type:varchar(255)" json:"remark"`                       // 备注说明
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Setting) TableName() string {
	return "settings"
}

// SettingGroup 设置分组常量
const (
	SettingGroupGeneral  = "general"  // 基础设置
	SettingGroupEmail    = "email"    // 邮件设置
	SettingGroupPayment  = "payment"  // 支付设置
	SettingGroupSecurity = "security" // 安全设置
	SettingGroupTheme    = "theme"    // 主题设置
	SettingGroupNode     = "node"     // 节点设置
	SettingGroupSubscribe = "subscribe" // 订阅设置
)

// 常用设置键名常量
const (
	// 基础设置
	SettingKeyAppName        = "app_name"
	SettingKeyAppURL         = "app_url"
	SettingKeyAppDescription = "app_description"
	SettingKeyAppLogo        = "app_logo"
	SettingKeyAppKeywords    = "app_keywords"
	SettingKeyTOSURL         = "tos_url"

	// 订阅设置
	SettingKeySubscribePath     = "subscribe_path"
	SettingKeySubscribeDomain   = "subscribe_domain"

	// 邮件设置
	SettingKeyEmailEnabled    = "email_enabled"
	SettingKeyEmailDriver     = "email_driver"     // smtp, api
	SettingKeySMTPHost        = "smtp_host"
	SettingKeySMTPPort        = "smtp_port"
	SettingKeySMTPUsername    = "smtp_username"
	SettingKeySMTPPassword    = "smtp_password"
	SettingKeySMTPFromAddress = "smtp_from_address"
	SettingKeySMTPFromName    = "smtp_from_name"
	SettingKeySMTPEncryption  = "smtp_encryption" // tls, ssl
	SettingKeyEmailAPIProvider = "email_api_provider" // sendgrid, mailgun, ses
	SettingKeyEmailAPIKey      = "email_api_key"

	// 支付设置
	SettingKeyPaymentEnabled     = "payment_enabled"
	SettingKeyAlipayEnabled      = "alipay_enabled"
	SettingKeyAlipayAppID        = "alipay_app_id"
	SettingKeyAlipayPrivateKey   = "alipay_private_key"
	SettingKeyAlipayPublicKey    = "alipay_public_key"
	SettingKeyWechatEnabled      = "wechat_enabled"
	SettingKeyWechatMchID        = "wechat_mch_id"
	SettingKeyWechatAPIKey       = "wechat_api_key"
	SettingKeyWechatCertPath     = "wechat_cert_path"
	SettingKeyWechatKeyPath      = "wechat_key_path"

	// 安全设置
	SettingKeySecurityRegisterLimit   = "security_register_limit"
	SettingKeySecurityLoginRetry      = "security_login_retry"
	SettingKeySecurity2FAEnabled      = "security_2fa_enabled"
	SettingKeySecurityIPWhitelist     = "security_ip_whitelist"
	SettingKeySecurityIPBlacklist     = "security_ip_blacklist"

	// 主题设置
	SettingKeyThemeFrontend = "frontend_theme"
	SettingKeyThemeAdmin    = "admin_theme"
	SettingKeyThemeSidebar  = "theme_sidebar"
	SettingKeyThemeHeader   = "theme_header"
	SettingKeyThemeColor    = "theme_color"
	SettingKeyBackgroundURL = "background_url"
)
