package service

import (
	"errors"
	"time"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
	"xboard-go/pkg/database"

	"gorm.io/gorm"
)

// SessionService 会话管理服务接口
type SessionService interface {
	GetActiveSessions(userID uint) ([]SessionInfo, error)
	RemoveSession(userID uint, sessionID string) error
}

// SessionInfo 会话信息
type SessionInfo struct {
	ID        string    `json:"id"`
	UserID    uint      `json:"user_id"`
	Token     string    `json:"token"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	IsCurrent bool      `json:"is_current"`
}

type sessionService struct {
	userRepo repository.UserRepository
	db       *gorm.DB
}

// NewSessionService 创建会话管理服务
func NewSessionService(userRepo repository.UserRepository) SessionService {
	return &sessionService{
		userRepo: userRepo,
		db:       database.Get(),
	}
}

// GetActiveSessions 获取用户活跃会话
func (s *sessionService) GetActiveSessions(userID uint) ([]SessionInfo, error) {
	// 查询有效的 token 记录
	var tokens []model.UserToken
	if err := s.db.Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&tokens).Error; err != nil {
		return nil, err
	}

	var sessions []SessionInfo
	for _, token := range tokens {
		sessions = append(sessions, SessionInfo{
			ID:        token.Token,
			UserID:    token.UserID,
			Token:     token.Token[:8] + "...", // 脱敏显示
			IP:        token.IP,
			UserAgent: token.UserAgent,
			ExpiresAt: token.ExpiresAt,
			CreatedAt: token.CreatedAt,
		})
	}

	if sessions == nil {
		sessions = []SessionInfo{}
	}

	return sessions, nil
}

// RemoveSession 移除指定会话
func (s *sessionService) RemoveSession(userID uint, sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}

	// 删除指定 token
	result := s.db.Where("user_id = ? AND token = ?", userID, sessionID).Delete(&model.UserToken{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("session not found")
	}

	return nil
}
