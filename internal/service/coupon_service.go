package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// CouponService 优惠券服务接口
type CouponService interface {
	Create(code, name string, couponType int, value, minAmount, maxDiscount float64, planIDs, userIDs string, limitCount int, startDate, endDate *time.Time) (*model.Coupon, error)
	GetByID(id uint) (*model.Coupon, error)
	GetByCode(code string) (*model.Coupon, error)
	Update(id uint, code, name string, couponType int, value, minAmount, maxDiscount float64, planIDs, userIDs string, limitCount int, startDate, endDate *time.Time, status int) (*model.Coupon, error)
	Delete(id uint) error
	List(page, pageSize int) ([]model.Coupon, int64, error)
	Validate(code string, userID uint, planID uint, amount float64) (float64, error)
	Use(id uint) error
}

type couponService struct {
	couponRepo repository.CouponRepository
}

// NewCouponService 创建优惠券服务
func NewCouponService(couponRepo repository.CouponRepository) CouponService {
	return &couponService{
		couponRepo: couponRepo,
	}
}

// Create 创建优惠券
func (s *couponService) Create(code, name string, couponType int, value, minAmount, maxDiscount float64, planIDs, userIDs string, limitCount int, startDate, endDate *time.Time) (*model.Coupon, error) {
	// 检查优惠码是否已存在
	existing, err := s.couponRepo.GetByCode(code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("coupon code already exists")
	}

	coupon := &model.Coupon{
		Code:        strings.ToUpper(code),
		Name:        name,
		Type:        couponType,
		Value:       value,
		MinAmount:   minAmount,
		MaxDiscount: maxDiscount,
		PlanIDs:     planIDs,
		UserIDs:     userIDs,
		LimitCount:  limitCount,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      1,
	}

	if err := s.couponRepo.Create(coupon); err != nil {
		return nil, err
	}

	return coupon, nil
}

// GetByID 根据ID获取优惠券
func (s *couponService) GetByID(id uint) (*model.Coupon, error) {
	coupon, err := s.couponRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if coupon == nil {
		return nil, errors.New("coupon not found")
	}
	return coupon, nil
}

// GetByCode 根据优惠码获取优惠券
func (s *couponService) GetByCode(code string) (*model.Coupon, error) {
	coupon, err := s.couponRepo.GetByCode(strings.ToUpper(code))
	if err != nil {
		return nil, err
	}
	if coupon == nil {
		return nil, errors.New("coupon not found")
	}
	return coupon, nil
}

// Update 更新优惠券
func (s *couponService) Update(id uint, code, name string, couponType int, value, minAmount, maxDiscount float64, planIDs, userIDs string, limitCount int, startDate, endDate *time.Time, status int) (*model.Coupon, error) {
	coupon, err := s.couponRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if coupon == nil {
		return nil, errors.New("coupon not found")
	}

	// 如果修改了优惠码，检查新码是否已存在
	if code != coupon.Code {
		existing, err := s.couponRepo.GetByCode(code)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, errors.New("coupon code already exists")
		}
	}

	coupon.Code = strings.ToUpper(code)
	coupon.Name = name
	coupon.Type = couponType
	coupon.Value = value
	coupon.MinAmount = minAmount
	coupon.MaxDiscount = maxDiscount
	coupon.PlanIDs = planIDs
	coupon.UserIDs = userIDs
	coupon.LimitCount = limitCount
	coupon.StartDate = startDate
	coupon.EndDate = endDate
	coupon.Status = status

	if err := s.couponRepo.Update(coupon); err != nil {
		return nil, err
	}

	return coupon, nil
}

// Delete 删除优惠券
func (s *couponService) Delete(id uint) error {
	coupon, err := s.couponRepo.GetByID(id)
	if err != nil {
		return err
	}
	if coupon == nil {
		return errors.New("coupon not found")
	}

	return s.couponRepo.Delete(id)
}

// List 优惠券列表
func (s *couponService) List(page, pageSize int) ([]model.Coupon, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.couponRepo.List(page, pageSize)
}

// Validate 验证优惠券
func (s *couponService) Validate(code string, userID uint, planID uint, amount float64) (float64, error) {
	coupon, err := s.couponRepo.GetByCode(strings.ToUpper(code))
	if err != nil {
		return 0, err
	}
	if coupon == nil {
		return 0, errors.New("coupon not found")
	}

	if !coupon.IsActive() {
		return 0, errors.New("coupon is not active")
	}

	// 检查用户是否可用
	if coupon.UserIDs != "" {
		userIDs := strings.Split(coupon.UserIDs, ",")
		found := false
		userIDStr := strconv.FormatUint(uint64(userID), 10)
		for _, id := range userIDs {
			if id == userIDStr {
				found = true
				break
			}
		}
		if !found {
			return 0, errors.New("coupon not available for this user")
		}
	}

	// 检查套餐是否可用
	if coupon.PlanIDs != "" {
		planIDs := strings.Split(coupon.PlanIDs, ",")
		found := false
		planIDStr := strconv.FormatUint(uint64(planID), 10)
		for _, id := range planIDs {
			if id == planIDStr {
				found = true
				break
			}
		}
		if !found {
			return 0, errors.New("coupon not available for this plan")
		}
	}

	discount := coupon.CalculateDiscount(amount)
	if discount <= 0 {
		return 0, errors.New("coupon cannot be applied")
	}

	return discount, nil
}

// Use 使用优惠券
func (s *couponService) Use(id uint) error {
	return s.couponRepo.IncrementUsedCount(id)
}
