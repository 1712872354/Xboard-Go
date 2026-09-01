package service

import (
	"fmt"
	"time"

	"xboard-go/config"
	"xboard-go/internal/model"
	"xboard-go/internal/repository"
	bizerrors "xboard-go/pkg/errors"
	"xboard-go/pkg/utils"
)

// UserService 用户服务接口
type UserService interface {
	GetUserByID(id uint) (*model.User, error)
	UpdateProfile(id uint, email string) (*model.User, error)
	ChangePassword(id uint, oldPassword, newPassword string) error
	ResetSubscribeToken(id uint) (string, error)
	ListUsers(page, pageSize int, filter repository.UserFilter) ([]model.User, int64, error)
	UpdateUserStatus(id uint, status int) error
	UpdateUserRole(id uint, role string) error
	DeleteUser(id uint) error
	AddTraffic(id uint, traffic int64, durationDays int) error
	// 管理员更新用户信息
	AdminUpdateUser(id uint, email string, password string, trafficLimit *int64, expiredAt *string, planID *uint, balance *float64, commission *float64) (*model.User, error)
	// 批量生成用户
	GenerateUsers(count int, prefix string, password string, planID int, expiredAt string) ([]model.User, error)
	// 导出用户CSV
	ExportUsersCSV(page, pageSize int, keyword string) (string, error)
}

type userService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(id uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, bizerrors.ErrUserNotFound
	}
	return user, nil
}

// UpdateProfile 更新用户资料
func (s *userService) UpdateProfile(id uint, email string) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, bizerrors.ErrUserNotFound
	}

	// 如果修改邮箱，检查是否已被使用
	if email != "" && email != user.Email {
		existingUser, err := s.userRepo.GetByEmail(email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil {
			return nil, bizerrors.ErrEmailExists
		}
		user.Email = email
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(id uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return bizerrors.ErrUserNotFound
	}

	// 验证旧密码
	if !utils.CheckPassword(oldPassword, user.PasswordHash) {
		return bizerrors.ErrInvalidPassword
	}

	// 加密新密码
	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	return s.userRepo.Update(user)
}

// GenerateUsers 批量生成用户
func (s *userService) GenerateUsers(count int, prefix string, password string, planID int, expiredAt string) ([]model.User, error) {
	if count <= 0 || count > 1000 {
		return nil, fmt.Errorf("count must be between 1 and 1000")
	}

	if prefix == "" {
		prefix = "user"
	}

	if password == "" {
		password = "123456" // 默认密码
	}

	// 加密密码
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var users []model.User
	for i := 0; i < count; i++ {
		// 生成邮箱
		email := fmt.Sprintf("%s%d@example.com", prefix, i+1)

		// 检查邮箱是否已存在
		existingUser, err := s.userRepo.GetByEmail(email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		if existingUser != nil {
			continue // 跳过已存在的邮箱
		}

		// 生成订阅token
		subscribeToken, err := utils.GenerateRandomString(config.Get().App.SubscribeTokenLength)
		if err != nil {
			return nil, fmt.Errorf("failed to generate subscribe token: %w", err)
		}

		// 创建用户
		user := &model.User{
			Email:          email,
			PasswordHash:   passwordHash,
			Role:           config.Get().App.DefaultUserRole,
			Status:         1,
			TrafficLimit:   0,
			UsedTraffic:    0,
			SubscribeToken: subscribeToken,
		}

		if err := s.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		users = append(users, *user)
	}

	return users, nil
}

// ExportUsersCSV 导出用户CSV
func (s *userService) ExportUsersCSV(page, pageSize int, keyword string) (string, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	users, _, err := s.userRepo.List(page, pageSize, repository.UserFilter{Keyword: keyword})
	if err != nil {
		return "", err
	}

	// 生成CSV内容
	csv := "ID,Email,Role,Status,TrafficLimit,UsedTraffic,ExpiredAt,CreatedAt\n"
	for _, user := range users {
		expiredAt := ""
		if user.ExpiredAt != nil {
			expiredAt = user.ExpiredAt.Format("2006-01-02 15:04:05")
		}
		csv += fmt.Sprintf("%d,%s,%s,%d,%d,%d,%s,%s\n",
			user.ID,
			user.Email,
			user.Role,
			user.Status,
			user.TrafficLimit,
			user.UsedTraffic,
			expiredAt,
			user.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}

	return csv, nil
}

// ResetSubscribeToken 重置订阅token
func (s *userService) ResetSubscribeToken(id uint) (string, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", bizerrors.ErrUserNotFound
	}

	newToken, err := utils.GenerateRandomString(config.Get().App.SubscribeTokenLength)
	if err != nil {
		return "", err
	}

	user.SubscribeToken = newToken
	if err := s.userRepo.Update(user); err != nil {
		return "", err
	}

	return newToken, nil
}

// ListUsers 用户列表（管理员）
func (s *userService) ListUsers(page, pageSize int, filter repository.UserFilter) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.userRepo.List(page, pageSize, filter)
}

// UpdateUserStatus 更新用户状态
func (s *userService) UpdateUserStatus(id uint, status int) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return bizerrors.ErrUserNotFound
	}

	user.Status = status
	return s.userRepo.Update(user)
}

// UpdateUserRole 更新用户角色
func (s *userService) UpdateUserRole(id uint, role string) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return bizerrors.ErrUserNotFound
	}

	user.Role = role
	return s.userRepo.Update(user)
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(id uint) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return bizerrors.ErrUserNotFound
	}

	return s.userRepo.Delete(id)
}

// AddTraffic 为用户增加流量和时长（购买套餐后调用）
func (s *userService) AddTraffic(id uint, traffic int64, durationDays int) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return bizerrors.ErrUserNotFound
	}

	// 增加流量
	user.TrafficLimit += traffic

	// 更新到期时间
	now := time.Now()
	if user.ExpiredAt == nil || user.ExpiredAt.Before(now) {
		// 如果没买过或已过期，从现在开始计算
		newExpiry := now.AddDate(0, 0, durationDays)
		user.ExpiredAt = &newExpiry
	} else {
		// 如果未过期，在现有基础上延长
		newExpiry := user.ExpiredAt.AddDate(0, 0, durationDays)
		user.ExpiredAt = &newExpiry
	}

	return s.userRepo.Update(user)
}

// AdminUpdateUser 管理员更新用户信息
func (s *userService) AdminUpdateUser(id uint, email string, password string, trafficLimit *int64, expiredAt *string, planID *uint, balance *float64, commission *float64) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, bizerrors.ErrUserNotFound
	}

	// 更新邮箱
	if email != "" && email != user.Email {
		existingUser, err := s.userRepo.GetByEmail(email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil {
			return nil, bizerrors.ErrEmailExists
		}
		user.Email = email
	}

	// 更新密码
	if password != "" {
		newHash, err := utils.HashPassword(password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = newHash
	}

	// 更新流量限制
	if trafficLimit != nil {
		user.TrafficLimit = *trafficLimit
	}

	// 更新到期时间
	if expiredAt != nil {
		if *expiredAt == "" {
			user.ExpiredAt = nil
		} else {
			t, err := time.Parse("2006-01-02 15:04:05", *expiredAt)
			if err != nil {
				t, err = time.Parse("2006-01-02", *expiredAt)
				if err != nil {
					return nil, fmt.Errorf("invalid date format: %w", err)
				}
			}
			user.ExpiredAt = &t
		}
	}

	// 更新套餐
	if planID != nil {
		if *planID == 0 {
			user.PlanID = nil
		} else {
			user.PlanID = planID
		}
	}

	// 更新余额
	if balance != nil {
		user.Balance = *balance
	}

	// 更新佣金
	if commission != nil {
		user.Commission = *commission
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}
