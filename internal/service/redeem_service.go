package service

import (
	"errors"
	"fmt"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
	"xboard-go/pkg/utils"
)

// RedeemService 兑换码服务接口
type RedeemService interface {
	// Generate 批量生成兑换码
	Generate(planID uint, count int, prefix string) ([]model.RedeemCode, error)
	// GetByID 根据ID获取兑换码
	GetByID(id uint) (*model.RedeemCode, error)
	// Redeem 兑换码兑换
	Redeem(code string, userID uint) (*model.Plan, error)
	// List 兑换码列表
	List(page, pageSize int, status int, planID uint) ([]model.RedeemCode, int64, error)
	// Delete 删除兑换码
	Delete(id uint) error
	// GetStats 获取兑换码统计
	GetStats() (*RedeemStats, error)
}

type redeemService struct {
	redeemCodeRepo repository.RedeemCodeRepository
	planRepo       repository.PlanRepository
	userService    UserService
}

// RedeemStats 兑换码统计
type RedeemStats struct {
	Total   int64 `json:"total"`
	Unused  int64 `json:"unused"`
	Used    int64 `json:"used"`
}

// NewRedeemService 创建兑换码服务
func NewRedeemService(
	redeemCodeRepo repository.RedeemCodeRepository,
	planRepo repository.PlanRepository,
	userService UserService,
) RedeemService {
	return &redeemService{
		redeemCodeRepo: redeemCodeRepo,
		planRepo:       planRepo,
		userService:    userService,
	}
}

// Generate 批量生成兑换码
func (s *redeemService) Generate(planID uint, count int, prefix string) ([]model.RedeemCode, error) {
	if count <= 0 || count > 1000 {
		return nil, errors.New("count must be between 1 and 1000")
	}

	// 验证套餐是否存在
	plan, err := s.planRepo.GetByID(planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}
	if plan == nil {
		return nil, errors.New("plan not found")
	}

	var codes []model.RedeemCode
	codeMap := make(map[string]bool)

	for i := 0; i < count; i++ {
		var codeStr string
		// 确保兑换码唯一
		for {
			random, err := utils.GenerateRandomString(12)
			if err != nil {
				return nil, fmt.Errorf("failed to generate code: %w", err)
			}
			codeStr = prefix + random
			if !codeMap[codeStr] {
				codeMap[codeStr] = true
				break
			}
		}

		codes = append(codes, model.RedeemCode{
			Code:   codeStr,
			PlanID: planID,
			Status: 0,
		})
	}

	if err := s.redeemCodeRepo.BatchCreate(codes); err != nil {
		return nil, fmt.Errorf("failed to batch create codes: %w", err)
	}

	return codes, nil
}

// GetByID 根据ID获取兑换码
func (s *redeemService) GetByID(id uint) (*model.RedeemCode, error) {
	code, err := s.redeemCodeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if code == nil {
		return nil, errors.New("redeem code not found")
	}
	return code, nil
}

// Redeem 兑换码兑换
func (s *redeemService) Redeem(code string, userID uint) (*model.Plan, error) {
	if code == "" {
		return nil, errors.New("redeem code is required")
	}

	// 检查用户是否存在
	user, err := s.userService.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 使用兑换码（事务内原子操作）
	redeemCode, err := s.redeemCodeRepo.Redeem(code, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to redeem code: %w", err)
	}

	// 获取套餐信息
	plan, err := s.planRepo.GetByID(redeemCode.PlanID)
	if err != nil {
		// 兑换码已使用，但套餐获取失败，记录日志（实际应补偿）
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	// 为用户增加流量和时长
	if err := s.userService.AddTraffic(userID, plan.Traffic, plan.DurationDays); err != nil {
		return nil, fmt.Errorf("failed to add traffic to user: %w", err)
	}

	return plan, nil
}

// List 兑换码列表
func (s *redeemService) List(page, pageSize int, status int, planID uint) ([]model.RedeemCode, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.redeemCodeRepo.List(page, pageSize, status, planID)
}

// Delete 删除兑换码
func (s *redeemService) Delete(id uint) error {
	code, err := s.redeemCodeRepo.GetByID(id)
	if err != nil {
		return err
	}
	if code == nil {
		return errors.New("redeem code not found")
	}
	if code.IsUsed() {
		return errors.New("cannot delete used redeem code")
	}
	return s.redeemCodeRepo.Delete(id)
}

// GetStats 获取兑换码统计
func (s *redeemService) GetStats() (*RedeemStats, error) {
	unused, err := s.redeemCodeRepo.CountByStatus(0)
	if err != nil {
		return nil, err
	}

	used, err := s.redeemCodeRepo.CountByStatus(1)
	if err != nil {
		return nil, err
	}

	return &RedeemStats{
		Total:  unused + used,
		Unused: unused,
		Used:   used,
	}, nil
}
