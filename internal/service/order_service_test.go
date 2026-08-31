package service

import (
	"testing"
	"time"

	"xboard-go/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderRepository 模拟订单仓储
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(order *model.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(id uint) (*model.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByTradeNo(tradeNo string) (*model.Order, error) {
	args := m.Called(tradeNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockOrderRepository) Update(order *model.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) ListByUserID(userID uint, page, pageSize int) ([]model.Order, int64, error) {
	args := m.Called(userID, page, pageSize)
	return args.Get(0).([]model.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) List(page, pageSize int, status int, userID uint) ([]model.Order, int64, error) {
	args := m.Called(page, pageSize, status, userID)
	return args.Get(0).([]model.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) UpdateStatus(orderID uint, status int, paidAt *time.Time) error {
	args := m.Called(orderID, status, paidAt)
	return args.Error(0)
}

// MockPlanRepository 模拟套餐仓储
type MockPlanRepository struct {
	mock.Mock
}

func (m *MockPlanRepository) Create(plan *model.Plan) error {
	args := m.Called(plan)
	return args.Error(0)
}

func (m *MockPlanRepository) GetByID(id uint) (*model.Plan, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Plan), args.Error(1)
}

func (m *MockPlanRepository) Update(plan *model.Plan) error {
	args := m.Called(plan)
	return args.Error(0)
}

func (m *MockPlanRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPlanRepository) List(page, pageSize int, includeDisabled bool) ([]model.Plan, int64, error) {
	args := m.Called(page, pageSize, includeDisabled)
	return args.Get(0).([]model.Plan), args.Get(1).(int64), args.Error(2)
}

func (m *MockPlanRepository) ListActive() ([]model.Plan, error) {
	args := m.Called()
	return args.Get(0).([]model.Plan), args.Error(1)
}

// MockCouponRepository 模拟优惠券仓储
type MockCouponRepository struct {
	mock.Mock
}

func (m *MockCouponRepository) Create(coupon *model.Coupon) error {
	args := m.Called(coupon)
	return args.Error(0)
}

func (m *MockCouponRepository) GetByID(id uint) (*model.Coupon, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetByCode(code string) (*model.Coupon, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Coupon), args.Error(1)
}

func (m *MockCouponRepository) Update(coupon *model.Coupon) error {
	args := m.Called(coupon)
	return args.Error(0)
}

func (m *MockCouponRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCouponRepository) List(page, pageSize int) ([]model.Coupon, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]model.Coupon), args.Get(1).(int64), args.Error(2)
}

func (m *MockCouponRepository) IncrementUsedCount(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// TestOrderService_CreateOrder 测试创建订单
func TestOrderService_CreateOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockPlanRepo := new(MockPlanRepository)
	mockUserRepo := new(MockUserRepository)
	mockCouponRepo := new(MockCouponRepository)
	orderService := NewOrderService(mockOrderRepo, mockPlanRepo, mockUserRepo, mockCouponRepo)

	t.Run("正常创建订单", func(t *testing.T) {
		// 模拟用户存在
		user := &model.User{
			ID:    1,
			Email: "test@example.com",
		}
		mockUserRepo.On("GetByID", uint(1)).Return(user, nil).Once()

		// 模拟套餐存在且可用
		plan := &model.Plan{
			ID:     1,
			Name:   "基础套餐",
			Price:  99.99,
			Status: 1,
		}
		mockPlanRepo.On("GetByID", uint(1)).Return(plan, nil).Once()

		// 模拟订单创建成功，并设置订单ID
		mockOrderRepo.On("Create", mock.AnythingOfType("*model.Order")).Run(func(args mock.Arguments) {
			order := args.Get(0).(*model.Order)
			order.ID = 1 // 模拟数据库自增ID
		}).Return(nil).Once()

		// 模拟重新获取订单
		createdOrder := &model.Order{
			ID:     1,
			UserID: 1,
			PlanID: 1,
			Amount: 99.99,
			Status: model.OrderStatusPending,
			Plan:   *plan,
		}
		mockOrderRepo.On("GetByID", uint(1)).Return(createdOrder, nil).Once()

		// 执行创建订单
		order, err := orderService.CreateOrder(1, 1, "")
		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, uint(1), order.UserID)
		assert.Equal(t, uint(1), order.PlanID)
		assert.Equal(t, 99.99, order.Amount)
		assert.Equal(t, model.OrderStatusPending, order.Status)
		mockOrderRepo.AssertExpectations(t)
		mockPlanRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("用户不存在", func(t *testing.T) {
		mockUserRepo.On("GetByID", uint(999)).Return(nil, nil).Once()

		order, err := orderService.CreateOrder(999, 1, "")
		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t, "user not found", err.Error())
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("套餐不存在", func(t *testing.T) {
		user := &model.User{
			ID:    1,
			Email: "test@example.com",
		}
		mockUserRepo.On("GetByID", uint(1)).Return(user, nil).Once()
		mockPlanRepo.On("GetByID", uint(999)).Return(nil, nil).Once()

		order, err := orderService.CreateOrder(1, 999, "")
		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t, "plan not found", err.Error())
		mockUserRepo.AssertExpectations(t)
		mockPlanRepo.AssertExpectations(t)
	})

	t.Run("套餐不可用", func(t *testing.T) {
		user := &model.User{
			ID:    1,
			Email: "test@example.com",
		}
		mockUserRepo.On("GetByID", uint(1)).Return(user, nil).Once()

		plan := &model.Plan{
			ID:     1,
			Name:   "下架套餐",
			Price:  99.99,
			Status: 0, // 下架
		}
		mockPlanRepo.On("GetByID", uint(1)).Return(plan, nil).Once()

		order, err := orderService.CreateOrder(1, 1, "")
		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t, "plan is not available", err.Error())
		mockUserRepo.AssertExpectations(t)
		mockPlanRepo.AssertExpectations(t)
	})
}
