package service

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// InviteService 邀请码服务接口
type InviteService interface {
	// 邀请码管理
	Create(userID uint, commission float64, limitCount int) (*model.InviteCode, error)
	GetByID(id uint) (*model.InviteCode, error)
	GetByCode(code string) (*model.InviteCode, error)
	GetByUserID(userID uint) (*model.InviteCode, error)
	Update(id uint, commission float64, limitCount int, status int) (*model.InviteCode, error)
	Delete(id uint) error
	List(page, pageSize int) ([]model.InviteCode, int64, error)

	// 使用邀请码
	UseCode(code string, userID uint) error

	// 佣金管理
	CreateCommissionLog(userID, fromUserID, orderID uint, amount, orderAmount, commission float64) (*model.CommissionLog, error)
	GetCommissionLogs(page, pageSize int, userID uint) ([]model.CommissionLog, int64, error)
	GetTotalCommission(userID uint) (float64, error)
	GetPendingCommission(userID uint) (float64, error)
	SettleCommission(id uint) error
	// 佣金提现（转入余额）
	WithdrawCommission(userID uint) (float64, error)
}

type inviteService struct {
	inviteRepo     repository.InviteCodeRepository
	commissionRepo repository.CommissionLogRepository
	userRepo       repository.UserRepository
}

// NewInviteService 创建邀请码服务
func NewInviteService(
	inviteRepo repository.InviteCodeRepository,
	commissionRepo repository.CommissionLogRepository,
	userRepo repository.UserRepository,
) InviteService {
	return &inviteService{
		inviteRepo:     inviteRepo,
		commissionRepo: commissionRepo,
		userRepo:       userRepo,
	}
}

// Create 创建邀请码
func (s *inviteService) Create(userID uint, commission float64, limitCount int) (*model.InviteCode, error) {
	// 检查用户是否已有邀请码
	existing, err := s.inviteRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("user already has an invite code")
	}

	code := &model.InviteCode{
		UserID:     userID,
		Code:       generateInviteCode(),
		Commission: commission,
		LimitCount: limitCount,
		Status:     1,
	}

	if err := s.inviteRepo.Create(code); err != nil {
		return nil, err
	}

	return code, nil
}

// generateInviteCode 生成邀请码
func generateInviteCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var code strings.Builder
	for i := 0; i < 8; i++ {
		code.WriteByte(chars[r.Intn(len(chars))])
	}

	return code.String()
}

// GetByID 根据ID获取邀请码
func (s *inviteService) GetByID(id uint) (*model.InviteCode, error) {
	code, err := s.inviteRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if code == nil {
		return nil, errors.New("invite code not found")
	}
	return code, nil
}

// GetByCode 根据邀请码获取
func (s *inviteService) GetByCode(code string) (*model.InviteCode, error) {
	inviteCode, err := s.inviteRepo.GetByCode(strings.ToUpper(code))
	if err != nil {
		return nil, err
	}
	if inviteCode == nil {
		return nil, errors.New("invite code not found")
	}
	return inviteCode, nil
}

// GetByUserID 根据用户ID获取邀请码
func (s *inviteService) GetByUserID(userID uint) (*model.InviteCode, error) {
	code, err := s.inviteRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if code == nil {
		return nil, errors.New("invite code not found")
	}
	return code, nil
}

// Update 更新邀请码
func (s *inviteService) Update(id uint, commission float64, limitCount int, status int) (*model.InviteCode, error) {
	code, err := s.inviteRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if code == nil {
		return nil, errors.New("invite code not found")
	}

	code.Commission = commission
	code.LimitCount = limitCount
	code.Status = status

	if err := s.inviteRepo.Update(code); err != nil {
		return nil, err
	}

	return code, nil
}

// Delete 删除邀请码
func (s *inviteService) Delete(id uint) error {
	code, err := s.inviteRepo.GetByID(id)
	if err != nil {
		return err
	}
	if code == nil {
		return errors.New("invite code not found")
	}

	return s.inviteRepo.Delete(id)
}

// List 邀请码列表
func (s *inviteService) List(page, pageSize int) ([]model.InviteCode, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.inviteRepo.List(page, pageSize)
}

// UseCode 使用邀请码
func (s *inviteService) UseCode(code string, userID uint) error {
	// 获取邀请码
	inviteCode, err := s.inviteRepo.GetByCode(strings.ToUpper(code))
	if err != nil {
		return err
	}
	if inviteCode == nil {
		return errors.New("invite code not found")
	}

	// 检查邀请码是否可用
	if !inviteCode.IsActive() {
		return errors.New("invite code is not active")
	}

	// 检查是否是自己的邀请码
	if inviteCode.UserID == userID {
		return errors.New("cannot use your own invite code")
	}

	// 增加使用次数
	if err := s.inviteRepo.IncrementUsedCount(inviteCode.ID); err != nil {
		return err
	}

	// 更新用户的邀请码ID
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	user.InviteCodeID = &inviteCode.ID
	return s.userRepo.Update(user)
}

// CreateCommissionLog 创建佣金记录
func (s *inviteService) CreateCommissionLog(userID, fromUserID, orderID uint, amount, orderAmount, commission float64) (*model.CommissionLog, error) {
	log := &model.CommissionLog{
		UserID:      userID,
		FromUserID:  fromUserID,
		OrderID:     orderID,
		Amount:      amount,
		OrderAmount: orderAmount,
		Commission:  commission,
		Status:      model.CommissionStatusPending,
	}

	if err := s.commissionRepo.Create(log); err != nil {
		return nil, err
	}

	return log, nil
}

// GetCommissionLogs 佣金记录列表
func (s *inviteService) GetCommissionLogs(page, pageSize int, userID uint) ([]model.CommissionLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.commissionRepo.List(page, pageSize, userID)
}

// GetTotalCommission 获取用户总佣金
func (s *inviteService) GetTotalCommission(userID uint) (float64, error) {
	return s.commissionRepo.GetTotalByUser(userID)
}

// GetPendingCommission 获取用户待结算佣金
func (s *inviteService) GetPendingCommission(userID uint) (float64, error) {
	return s.commissionRepo.GetPendingByUser(userID)
}

// SettleCommission 结算佣金
func (s *inviteService) SettleCommission(id uint) error {
	log, err := s.commissionRepo.GetByID(id)
	if err != nil {
		return err
	}
	if log == nil {
		return errors.New("commission log not found")
	}

	if log.Status != model.CommissionStatusPending {
		return errors.New("commission is not pending")
	}

	// 更新状态为已结算
	log.Status = model.CommissionStatusSettled
	if err := s.commissionRepo.Update(log); err != nil {
		return err
	}

	// 增加用户佣金余额
	user, err := s.userRepo.GetByID(log.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	user.Commission += log.Amount
	return s.userRepo.Update(user)
}

// WithdrawCommission 佣金提现（将佣金转入余额）
func (s *inviteService) WithdrawCommission(userID uint) (float64, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, errors.New("user not found")
	}

	// 获取待结算佣金
	pending, err := s.commissionRepo.GetPendingByUser(userID)
	if err != nil {
		return 0, err
	}

	if pending <= 0 {
		return 0, errors.New("no pending commission to withdraw")
	}

	// 将待结算佣金转入余额
	user.Balance += pending
	user.Commission = 0 // 清空佣金余额

	if err := s.userRepo.Update(user); err != nil {
		return 0, err
	}

	// 将所有待结算记录标记为已结算
	logs, _, err := s.commissionRepo.List(1, 1000, userID)
	if err == nil {
		for _, log := range logs {
			if log.Status == model.CommissionStatusPending {
				log.Status = model.CommissionStatusSettled
				s.commissionRepo.Update(&log)
			}
		}
	}

	return pending, nil
}
