package service

import (
	"errors"
	"fmt"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// PlanService 套餐服务接口
type PlanService interface {
	CreatePlan(name string, price float64, traffic int64, durationDays int, deviceLimit int, nodeGroup, description string) (*model.Plan, error)
	GetPlanByID(id uint) (*model.Plan, error)
	UpdatePlan(id uint, name string, price float64, traffic int64, durationDays int, deviceLimit int, nodeGroup, description string, status int) (*model.Plan, error)
	DeletePlan(id uint) error
	ListPlans(page, pageSize int, includeDisabled bool) ([]model.Plan, int64, error)
	ListActivePlans() ([]model.Plan, error)
}

type planService struct {
	planRepo repository.PlanRepository
}

// NewPlanService 创建套餐服务
func NewPlanService(planRepo repository.PlanRepository) PlanService {
	return &planService{
		planRepo: planRepo,
	}
}

// CreatePlan 创建套餐
func (s *planService) CreatePlan(name string, price float64, traffic int64, durationDays int, deviceLimit int, nodeGroup, description string) (*model.Plan, error) {
	if name == "" {
		return nil, errors.New("plan name is required")
	}
	if price < 0 {
		return nil, errors.New("price must be non-negative")
	}
	if traffic < 0 {
		return nil, errors.New("traffic must be non-negative")
	}
	if durationDays <= 0 {
		return nil, errors.New("duration days must be positive")
	}

	plan := &model.Plan{
		Name:         name,
		Price:        price,
		Traffic:      traffic,
		DurationDays: durationDays,
		DeviceLimit:  deviceLimit,
		NodeGroup:    nodeGroup,
		Description:  description,
		Status:       1,
	}

	if err := s.planRepo.Create(plan); err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	return plan, nil
}

// GetPlanByID 根据ID获取套餐
func (s *planService) GetPlanByID(id uint) (*model.Plan, error) {
	plan, err := s.planRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, errors.New("plan not found")
	}
	return plan, nil
}

// UpdatePlan 更新套餐
func (s *planService) UpdatePlan(id uint, name string, price float64, traffic int64, durationDays int, deviceLimit int, nodeGroup, description string, status int) (*model.Plan, error) {
	plan, err := s.planRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, errors.New("plan not found")
	}

	if name != "" {
		plan.Name = name
	}
	if price >= 0 {
		plan.Price = price
	}
	if traffic >= 0 {
		plan.Traffic = traffic
	}
	if durationDays > 0 {
		plan.DurationDays = durationDays
	}
	if deviceLimit >= 0 {
		plan.DeviceLimit = deviceLimit
	}
	if nodeGroup != "" {
		plan.NodeGroup = nodeGroup
	}
	if description != "" {
		plan.Description = description
	}
	if status == 0 || status == 1 {
		plan.Status = status
	}

	if err := s.planRepo.Update(plan); err != nil {
		return nil, fmt.Errorf("failed to update plan: %w", err)
	}

	return plan, nil
}

// DeletePlan 删除套餐
func (s *planService) DeletePlan(id uint) error {
	plan, err := s.planRepo.GetByID(id)
	if err != nil {
		return err
	}
	if plan == nil {
		return errors.New("plan not found")
	}

	return s.planRepo.Delete(id)
}

// ListPlans 套餐列表（管理员）
func (s *planService) ListPlans(page, pageSize int, includeDisabled bool) ([]model.Plan, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.planRepo.List(page, pageSize, includeDisabled)
}

// ListActivePlans 获取上架的套餐列表（用户可见）
func (s *planService) ListActivePlans() ([]model.Plan, error) {
	return s.planRepo.ListActive()
}
