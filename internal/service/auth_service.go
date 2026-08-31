package service

import (
	"fmt"
	"time"

	"xboard-go/config"
	"xboard-go/internal/model"
	"xboard-go/internal/repository"
	"xboard-go/pkg/database"
	bizerrors "xboard-go/pkg/errors"
	"xboard-go/pkg/jwt"
	"xboard-go/pkg/logger"
	"xboard-go/pkg/utils"
)

// AuthService 认证服务接口
type AuthService interface {
	Register(email, password string) (*model.User, error)
	Login(email, password string) (*jwt.TokenPair, *model.User, error)
	RefreshToken(refreshToken string) (*jwt.TokenPair, error)
	ForgetPassword(email string) error
	ResetPassword(token, newPassword string) error
	SendVerificationCode(email string) error
	VerifyEmail(email, code string) error
}

type authService struct {
	userRepo repository.UserRepository
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

// Register 用户注册
func (s *authService) Register(email, password string) (*model.User, error) {
	// 检查邮箱是否已存在
	existingUser, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existingUser != nil {
		return nil, bizerrors.ErrEmailExists
	}

	// 加密密码
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 生成订阅 token
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

	// 试用套餐处理
	go s.assignTrialPlan(user)

	return user, nil
}

// assignTrialPlan 分配套试用套餐
func (s *authService) assignTrialPlan(user *model.User) {
	// 安全获取数据库连接（测试环境可能未初始化）
	defer func() {
		if r := recover(); r != nil {
			logger.Sugar().Debugf("Trial plan assignment skipped: %v", r)
		}
	}()

	db := database.Get()
	if db == nil {
		return
	}

	// 读取试用套餐配置
	var trialPlanIDSetting, trialHourSetting model.Setting
	var trialPlanID, trialHour int

	if err := db.Where("`key` = ?", "try_out_plan_id").First(&trialPlanIDSetting).Error; err != nil {
		return // 未配置试用套餐
	}
	if _, err := fmt.Sscanf(trialPlanIDSetting.Value, "%d", &trialPlanID); err != nil || trialPlanID <= 0 {
		return
	}

	trialHour = 1 // 默认1小时
	if err := db.Where("`key` = ?", "try_out_hour").First(&trialHourSetting).Error; err == nil {
		fmt.Sscanf(trialHourSetting.Value, "%d", &trialHour)
	}
	if trialHour <= 0 {
		trialHour = 1
	}

	// 获取试用套餐
	var plan model.Plan
	if err := db.First(&plan, trialPlanID).Error; err != nil {
		logger.Sugar().Warnf("Trial plan %d not found: %v", trialPlanID, err)
		return
	}

	if !plan.IsAvailable() {
		return
	}

	// 分配套餐给用户
	now := time.Now()
	expiredAt := now.Add(time.Duration(trialHour) * time.Hour)
	planID := uint(trialPlanID)

	updates := map[string]interface{}{
		"plan_id":       planID,
		"traffic_limit": plan.Traffic,
		"expired_at":    expiredAt,
	}

	if err := db.Model(user).Updates(updates).Error; err != nil {
		logger.Sugar().Warnf("Failed to assign trial plan to user %d: %v", user.ID, err)
		return
	}

	// 创建试用订单
	order := &model.Order{
		UserID:      user.ID,
		PlanID:      uint(trialPlanID),
		TradeNo:     fmt.Sprintf("TRIAL%d%d", user.ID, now.Unix()),
		Amount:      0,
		Status:      model.OrderStatusPaid,
		PaidAt:      &now,
		CreatedAt:   now,
	}

	if err := db.Create(order).Error; err != nil {
		logger.Sugar().Warnf("Failed to create trial order for user %d: %v", user.ID, err)
	}

	logger.Sugar().Infof("Assigned trial plan %d to user %d for %d hours", trialPlanID, user.ID, trialHour)
}

// Login 用户登录
func (s *authService) Login(email, password string) (*jwt.TokenPair, *model.User, error) {
	// 获取用户
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, nil, bizerrors.ErrInvalidCredentials
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, nil, bizerrors.ErrAccountBanned
	}

	// 验证密码
	if !utils.CheckPassword(password, user.PasswordHash) {
		return nil, nil, bizerrors.ErrInvalidCredentials
	}

	// 生成 Token
	tokenPair, err := jwt.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenPair, user, nil
}

// RefreshToken 刷新 Token
func (s *authService) RefreshToken(refreshToken string) (*jwt.TokenPair, error) {
	// 解析token获取用户信息
	claims, err := jwt.ParseToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 检查用户状态
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, bizerrors.ErrUserNotFound
	}
	if !user.IsActive() {
		return nil, bizerrors.ErrAccountBanned
	}

	// 刷新token
	tokenPair, err := jwt.RefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	return tokenPair, nil
}

// ForgetPassword 忘记密码
func (s *authService) ForgetPassword(email string) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		// 为了安全，即使用户不存在也返回成功
		return nil
	}

	// 生成重置密码token
	_, err = utils.GenerateRandomString(32)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// TODO: 将token保存到数据库，并设置过期时间
	// TODO: 发送重置密码邮件

	return nil
}

// ResetPassword 重置密码
func (s *authService) ResetPassword(token, newPassword string) error {
	// TODO: 验证token的有效性
	// TODO: 从token中获取用户信息
	// TODO: 更新用户密码

	return fmt.Errorf("not implemented")
}

// SendVerificationCode 发送验证码
func (s *authService) SendVerificationCode(email string) error {
	// 生成验证码
	_, err := utils.GenerateRandomString(6)
	if err != nil {
		return fmt.Errorf("failed to generate verification code: %w", err)
	}

	// TODO: 将验证码保存到Redis，设置过期时间
	// TODO: 发送验证码邮件

	return nil
}

// VerifyEmail 验证邮箱
func (s *authService) VerifyEmail(email, code string) error {
	// TODO: 从Redis获取验证码
	// TODO: 验证验证码是否正确
	// TODO: 更新用户邮箱验证状态

	return fmt.Errorf("not implemented")
}
