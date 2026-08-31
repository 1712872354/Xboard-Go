package captcha

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Provider 验证码提供商
type Provider string

const (
	ProviderTurnstile  Provider = "turnstile"
	ProviderRecaptchaV2 Provider = "recaptcha_v2"
	ProviderRecaptchaV3 Provider = "recaptcha_v3"
)

// Config 验证码配置
type Config struct {
	Provider    Provider `mapstructure:"provider"`
	SiteKey     string   `mapstructure:"site_key"`
	SecretKey   string   `mapstructure:"secret_key"`
	MinScore    float64  `mapstructure:"min_score"` // reCAPTCHA v3 最小分数
}

// VerifyRequest 验证请求
type VerifyRequest struct {
	Token      string `json:"token"`
	RemoteIP   string `json:"remote_ip,omitempty"`
}

// VerifyResponse 验证响应
type VerifyResponse struct {
	Success     bool     `json:"success"`
	ErrorCodes []string `json:"error-codes,omitempty"`
	Score      float64  `json:"score,omitempty"`      // reCAPTCHA v3
	Action     string   `json:"action,omitempty"`      // reCAPTCHA v3
	ChallengeTS string  `json:"challenge_ts,omitempty"`
	Hostname   string   `json:"hostname,omitempty"`
}

// CaptchaService 验证码服务
type CaptchaService struct {
	config Config
	client *http.Client
}

// NewCaptchaService 创建验证码服务
func NewCaptchaService(config Config) *CaptchaService {
	return &CaptchaService{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Verify 验证验证码
func (s *CaptchaService) Verify(token, remoteIP string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("captcha token is required")
	}

	switch s.config.Provider {
	case ProviderTurnstile:
		return s.verifyTurnstile(token, remoteIP)
	case ProviderRecaptchaV2:
		return s.verifyRecaptcha(token, remoteIP)
	case ProviderRecaptchaV3:
		return s.verifyRecaptchaV3(token, remoteIP)
	default:
		return false, fmt.Errorf("unsupported captcha provider: %s", s.config.Provider)
	}
}

// verifyTurnstile 验证 Cloudflare Turnstile
func (s *CaptchaService) verifyTurnstile(token, remoteIP string) (bool, error) {
	form := url.Values{}
	form.Set("secret", s.config.SecretKey)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	resp, err := s.client.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", form)
	if err != nil {
		return false, fmt.Errorf("failed to verify turnstile: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	var result VerifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Success, nil
}

// verifyRecaptcha 验证 Google reCAPTCHA v2
func (s *CaptchaService) verifyRecaptcha(token, remoteIP string) (bool, error) {
	form := url.Values{}
	form.Set("secret", s.config.SecretKey)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	resp, err := s.client.PostForm("https://www.google.com/recaptcha/api/siteverify", form)
	if err != nil {
		return false, fmt.Errorf("failed to verify recaptcha: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	var result VerifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Success, nil
}

// verifyRecaptchaV3 验证 Google reCAPTCHA v3
func (s *CaptchaService) verifyRecaptchaV3(token, remoteIP string) (bool, error) {
	form := url.Values{}
	form.Set("secret", s.config.SecretKey)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	resp, err := s.client.PostForm("https://www.google.com/recaptcha/api/siteverify", form)
	if err != nil {
		return false, fmt.Errorf("failed to verify recaptcha v3: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	var result VerifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return false, nil
	}

	// 检查分数是否达到最小要求
	minScore := s.config.MinScore
	if minScore <= 0 {
		minScore = 0.5 // 默认最小分数
	}

	return result.Score >= minScore, nil
}

// GetSiteKey 获取站点密钥（用于前端）
func (s *CaptchaService) GetSiteKey() string {
	return s.config.SiteKey
}

// GetProvider 获取验证码提供商
func (s *CaptchaService) GetProvider() Provider {
	return s.config.Provider
}

// IsEnabled 是否启用了验证码
func (s *CaptchaService) IsEnabled() bool {
	return s.config.Provider != "" && s.config.SiteKey != "" && s.config.SecretKey != ""
}
