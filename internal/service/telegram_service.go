package service

import (
	"fmt"
	"strings"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/logger"
	"xboard-go/pkg/telegram"

	"gorm.io/gorm"
)

// TelegramService Telegram 服务接口
type TelegramService interface {
	// HandleWebhook 处理 Webhook 消息
	HandleWebhook(update *telegram.Update) error
	// SendAdminMessage 发送消息给管理员
	SendAdminMessage(message string) error
	// SendUserMessage 发送消息给用户
	SendUserMessage(chatID int64, message string) error
	// SetWebhook 设置 Webhook
	SetWebhook(webhookURL string) error
	// RegisterCommands 注册 Bot 命令
	RegisterCommands() error
	// GetUserByTelegramID 通过 Telegram ID 获取用户
	GetUserByTelegramID(telegramID int64) (*model.User, error)
	// BindUser 绑定用户
	BindUser(telegramID int64, subscribeToken string) error
	// UnbindUser 解绑用户
	UnbindUser(telegramID int64) error
}

type telegramService struct {
	db     *gorm.DB
	client *telegram.Client
}

// NewTelegramService 创建 Telegram 服务
func NewTelegramService() TelegramService {
	return &telegramService{
		db: database.Get(),
	}
}

// getClient 获取 Telegram 客户端（延迟初始化）
func (s *telegramService) getClient() (*telegram.Client, error) {
	if s.client != nil {
		return s.client, nil
	}

	token := s.getSetting("telegram_bot_token", "")
	if token == "" {
		return nil, fmt.Errorf("telegram bot token not configured")
	}

	s.client = telegram.NewClient(token)
	return s.client, nil
}

// getSetting 从数据库获取设置值
func (s *telegramService) getSetting(key, defaultValue string) string {
	var setting model.Setting
	if err := s.db.Where("`key` = ?", key).First(&setting).Error; err != nil {
		return defaultValue
	}
	if setting.Value == "" {
		return defaultValue
	}
	return setting.Value
}

// HandleWebhook 处理 Webhook 消息
func (s *telegramService) HandleWebhook(update *telegram.Update) error {
	if update.Message == nil {
		return nil
	}

	msg := update.Message
	chatID := msg.Chat.ID
	text := msg.Text

	// 处理命令
	if strings.HasPrefix(text, "/") {
		return s.handleCommand(chatID, text, msg)
	}

	// 处理回复消息（用于工单回复）
	if msg.ReplyToMessage != nil {
		return s.handleReplyMessage(chatID, msg)
	}

	return nil
}

// handleCommand 处理命令
func (s *telegramService) handleCommand(chatID int64, text string, msg *telegram.Message) error {
	parts := strings.SplitN(text, " ", 2)
	command := parts[0]
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	switch command {
	case "/start":
		return s.handleStartCommand(chatID)
	case "/bind":
		return s.handleBindCommand(chatID, args)
	case "/unbind":
		return s.handleUnbindCommand(chatID)
	case "/traffic":
		return s.handleTrafficCommand(chatID)
	case "/getlatesturl":
		return s.handleGetLatestURLCommand(chatID)
	default:
		return s.SendMessage(chatID, "未知命令，请输入 /start 查看帮助")
	}
}

// handleStartCommand 处理 /start 命令
func (s *telegramService) handleStartCommand(chatID int64) error {
	user, _ := s.GetUserByTelegramID(chatID)

	var text string
	if user != nil {
		text = fmt.Sprintf(`🎉 欢迎使用 %s Telegram Bot！

✅ 您已绑定账号：%s

📋 可用命令：
/traffic - 查看流量使用情况
/getlatesturl - 获取订阅链接
/unbind - 解绑账号`,
			s.getSetting("app_name", "XBoard"),
			user.Email)
	} else {
		text = fmt.Sprintf(`🎉 欢迎使用 %s Telegram Bot！

🔗 请先绑定您的账号：
1. 登录您的账户
2. 复制您的订阅链接
3. 发送 /bind + 订阅链接

📋 可用命令：
/bind [订阅链接] - 绑定账号`,
			s.getSetting("app_name", "XBoard"))
	}

	return s.SendMessage(chatID, text)
}

// handleBindCommand 处理 /bind 命令
func (s *telegramService) handleBindCommand(chatID int64, subscribeURL string) error {
	if subscribeURL == "" {
		return s.SendMessage(chatID, "参数有误，请携带订阅地址发送\n格式：/bind [订阅链接]")
	}

	// 检查是否已绑定
	existingUser, _ := s.GetUserByTelegramID(chatID)
	if existingUser != nil {
		return s.SendMessage(chatID, "您的账号已经绑定了 Telegram，无需重复绑定")
	}

	// 从 URL 提取 token
	token := extractTokenFromURL(subscribeURL)
	if token == "" {
		return s.SendMessage(chatID, "订阅地址无效")
	}

	// 查找用户
	var user model.User
	if err := s.db.Where("subscribe_token = ?", token).First(&user).Error; err != nil {
		return s.SendMessage(chatID, "用户不存在")
	}

	// 检查 token 是否已被绑定
	if user.TelegramID > 0 {
		return s.SendMessage(chatID, "该账号已经绑定了其他 Telegram 账号")
	}

	// 绑定
	if err := s.db.Model(&user).Update("telegram_id", chatID).Error; err != nil {
		logger.Sugar().Errorf("Failed to bind Telegram user: %v", err)
		return s.SendMessage(chatID, "绑定失败，请稍后重试")
	}

	return s.SendMessage(chatID, "绑定成功！")
}

// handleUnbindCommand 处理 /unbind 命令
func (s *telegramService) handleUnbindCommand(chatID int64) error {
	user, err := s.GetUserByTelegramID(chatID)
	if err != nil {
		return s.SendMessage(chatID, "请先绑定账号")
	}

	if err := s.db.Model(user).Update("telegram_id", 0).Error; err != nil {
		return s.SendMessage(chatID, "解绑失败")
	}

	return s.SendMessage(chatID, "解绑成功！")
}

// handleTrafficCommand 处理 /traffic 命令
func (s *telegramService) handleTrafficCommand(chatID int64) error {
	user, err := s.GetUserByTelegramID(chatID)
	if err != nil {
		return s.SendMessage(chatID, "请先绑定账号")
	}

	totalTraffic := user.TrafficLimit
	usedTraffic := user.UsedTraffic
	remaining := totalTraffic - usedTraffic
	if remaining < 0 {
		remaining = 0
	}

	var percentage float64
	if totalTraffic > 0 {
		percentage = float64(usedTraffic) / float64(totalTraffic) * 100
	}

	text := fmt.Sprintf(`📊 流量使用情况

已用流量：%s
总流量：%s
剩余流量：%s
使用率：%.2f%%`,
		telegram.EscapeMarkdown(FormatTraffic(usedTraffic)),
		telegram.EscapeMarkdown(FormatTraffic(totalTraffic)),
		telegram.EscapeMarkdown(FormatTraffic(remaining)),
		percentage)

	return s.SendMessage(chatID, text)
}

// handleGetLatestURLCommand 处理 /getlatesturl 命令
func (s *telegramService) handleGetLatestURLCommand(chatID int64) error {
	user, err := s.GetUserByTelegramID(chatID)
	if err != nil {
		return s.SendMessage(chatID, "请先绑定账号")
	}

	subscribeURL := s.getSetting("subscribe_url", "")
	if subscribeURL == "" {
		subscribeURL = s.getSetting("app_url", "")
	}
	if subscribeURL == "" {
		return s.SendMessage(chatID, "订阅地址未配置")
	}

	subscribeURL = strings.TrimRight(subscribeURL, "/") + "/api/v1/client/subscribe?token=" + user.SubscribeToken

	text := fmt.Sprintf("🔗 您的订阅链接：\n\n%s", subscribeURL)
	return s.SendMessage(chatID, text)
}

// handleReplyMessage 处理回复消息（用于工单回复）
func (s *telegramService) handleReplyMessage(chatID int64, msg *telegram.Message) error {
	// 检查是否是工单回复
	if msg.ReplyToMessage == nil {
		return nil
	}

	replyText := msg.ReplyToMessage.Text
	// 尝试从回复消息中提取工单ID
	// 格式：📮 *工单提醒* #123 或 工单ID: 123
	var ticketID uint
	if _, err := fmt.Sscanf(replyText, "📮 工单提醒 #%d", &ticketID); err != nil {
		// 尝试其他格式
		if _, err := fmt.Sscanf(replyText, "工单ID: %d", &ticketID); err != nil {
			return nil // 不是工单消息，忽略
		}
	}

	if ticketID == 0 {
		return nil
	}

	// 获取用户
	user, err := s.GetUserByTelegramID(chatID)
	if err != nil {
		return s.SendMessage(chatID, "请先绑定账号")
	}

	// 检查工单是否存在
	var ticket model.Ticket
	if err := s.db.First(&ticket, ticketID).Error; err != nil {
		return s.SendMessage(chatID, "工单不存在")
	}

	// 创建工单回复
	reply := model.TicketReply{
		TicketID: ticketID,
		UserID:   user.ID,
		Content:  msg.Text,
		IsAdmin:  user.IsAdmin(),
	}

	if err := s.db.Create(&reply).Error; err != nil {
		return s.SendMessage(chatID, "回复失败")
	}

	// 更新工单状态
	s.db.Model(&ticket).Updates(map[string]interface{}{
		"status":     model.TicketStatusReplied,
		"last_reply": gorm.Expr("CURRENT_TIMESTAMP"),
	})

	return s.SendMessage(chatID, fmt.Sprintf("工单 #%d 回复成功", ticketID))
}

// SendMessage 发送消息
func (s *telegramService) SendMessage(chatID int64, text string) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}
	return client.SendMessage(chatID, text, "Markdown")
}

// SendAdminMessage 发送消息给所有管理员
func (s *telegramService) SendAdminMessage(message string) error {
	var admins []model.User
	if err := s.db.Where("role = ? AND telegram_id > 0", "admin").Find(&admins).Error; err != nil {
		return err
	}

	for _, admin := range admins {
		if err := s.SendMessage(admin.TelegramID, message); err != nil {
			logger.Sugar().Warnf("Failed to send Telegram to admin %d: %v", admin.ID, err)
		}
	}
	return nil
}

// SendUserMessage 发送消息给用户
func (s *telegramService) SendUserMessage(chatID int64, message string) error {
	return s.SendMessage(chatID, message)
}

// SetWebhook 设置 Webhook
func (s *telegramService) SetWebhook(webhookURL string) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}
	return client.SetWebhook(webhookURL)
}

// RegisterCommands 注册 Bot 命令
func (s *telegramService) RegisterCommands() error {
	client, err := s.getClient()
	if err != nil {
		return err
	}

	commands := []telegram.BotCommand{
		{Command: "start", Description: "开始使用"},
		{Command: "bind", Description: "绑定账号"},
		{Command: "traffic", Description: "查看流量"},
		{Command: "getlatesturl", Description: "获取订阅链接"},
		{Command: "unbind", Description: "解绑账号"},
	}

	return client.SetMyCommands(commands)
}

// GetUserByTelegramID 通过 Telegram ID 获取用户
func (s *telegramService) GetUserByTelegramID(telegramID int64) (*model.User, error) {
	var user model.User
	if err := s.db.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// BindUser 绑定用户
func (s *telegramService) BindUser(telegramID int64, subscribeToken string) error {
	var user model.User
	if err := s.db.Where("subscribe_token = ?", subscribeToken).First(&user).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	if user.TelegramID > 0 {
		return fmt.Errorf("user already bound to another Telegram account")
	}

	return s.db.Model(&user).Update("telegram_id", telegramID).Error
}

// UnbindUser 解绑用户
func (s *telegramService) UnbindUser(telegramID int64) error {
	var user model.User
	if err := s.db.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	return s.db.Model(&user).Update("telegram_id", 0).Error
}

// extractTokenFromURL 从订阅 URL 提取 token
func extractTokenFromURL(rawURL string) string {
	// 尝试从查询参数提取
	if idx := strings.Index(rawURL, "token="); idx >= 0 {
		token := rawURL[idx+6:]
		if endIdx := strings.Index(token, "&"); endIdx >= 0 {
			token = token[:endIdx]
		}
		return token
	}

	// 尝试从路径提取（最后一个路径段）
	parts := strings.Split(strings.TrimRight(rawURL, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return ""
}
