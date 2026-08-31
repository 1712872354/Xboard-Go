package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const telegramAPIBase = "https://api.telegram.org/bot"

// Client Telegram Bot API 客户端
type Client struct {
	token      string
	httpClient *http.Client
	apiBase    string
}

// NewClient 创建 Telegram 客户端
func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiBase: telegramAPIBase + token + "/",
	}
}

// APIResponse Telegram API 响应
type APIResponse struct {
	OK          bool            `json:"ok"`
	Description string          `json:"description,omitempty"`
	Result      json.RawMessage `json:"result,omitempty"`
}

// User Telegram 用户
type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// Chat Telegram 聊天
type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// Message Telegram 消息
type Message struct {
	MessageID int    `json:"message_id"`
	From      *User  `json:"from,omitempty"`
	Chat      *Chat  `json:"chat"`
	Date      int64  `json:"date"`
	Text      string `json:"text,omitempty"`
	ReplyToMessage *Message `json:"reply_to_message,omitempty"`
}

// Update Telegram 更新
type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message,omitempty"`
}

// BotCommand Bot 命令
type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

// GetMe 获取机器人信息
func (c *Client) GetMe() (*User, error) {
	var user User
	if err := c.request("getMe", nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// SendMessage 发送消息
func (c *Client) SendMessage(chatID int64, text string, parseMode string) error {
	params := map[string]string{
		"chat_id": fmt.Sprintf("%d", chatID),
		"text":    text,
	}
	if parseMode != "" {
		params["parse_mode"] = parseMode
	}
	return c.request("sendMessage", params, nil)
}

// SetWebhook 设置 Webhook
func (c *Client) SetWebhook(webhookURL string) error {
	params := map[string]string{
		"url": webhookURL,
	}
	return c.request("setWebhook", params, nil)
}

// SetMyCommands 设置 Bot 命令列表
func (c *Client) SetMyCommands(commands []BotCommand) error {
	cmdJSON, err := json.Marshal(commands)
	if err != nil {
		return err
	}
	params := map[string]string{
		"commands": string(cmdJSON),
	}
	return c.request("setMyCommands", params, nil)
}

// GetMyCommands 获取 Bot 命令列表
func (c *Client) GetMyCommands() ([]BotCommand, error) {
	var commands []BotCommand
	if err := c.request("getMyCommands", nil, &commands); err != nil {
		return nil, err
	}
	return commands, nil
}

// DeleteMyCommands 删除所有命令
func (c *Client) DeleteMyCommands() error {
	return c.request("deleteMyCommands", nil, nil)
}

// request 发送 API 请求
func (c *Client) request(method string, params map[string]string, result interface{}) error {
	apiURL := c.apiBase + method

	form := url.Values{}
	for key, value := range params {
		form.Set(key, value)
	}

	var body io.Reader
	if len(params) > 0 {
		body = strings.NewReader(form.Encode())
	}

	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	if len(params) > 0 {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("parse response failed: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram API error: %s", apiResp.Description)
	}

	if result != nil && len(apiResp.Result) > 0 {
		if err := json.Unmarshal(apiResp.Result, result); err != nil {
			return fmt.Errorf("parse result failed: %w", err)
		}
	}

	return nil
}

// EscapeMarkdown 转义 Markdown 特殊字符
func EscapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}
