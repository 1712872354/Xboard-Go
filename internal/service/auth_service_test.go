package service

import (
	"os"
	"testing"

	"xboard-go/config"
	"xboard-go/internal/model"
	"xboard-go/internal/repository"
	bizerrors "xboard-go/pkg/errors"
	"xboard-go/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	// 初始化配置用于测试
	_, err := config.Load("../../config.test.yaml")
	if err != nil {
		panic("failed to load test config: " + err.Error())
	}
	os.Exit(m.Run())
}

// MockUserRepository 模拟用户仓储
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(id uint) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Create(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) GetBySubscribeToken(token string) (*model.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) UpdateTraffic(userID uint, traffic int64) error {
	args := m.Called(userID, traffic)
	return args.Error(0)
}

func (m *MockUserRepository) List(page, pageSize int, filter repository.UserFilter) ([]model.User, int64, error) {
	args := m.Called(page, pageSize, filter)
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

// TestAuthService_Register 测试用户注册
func TestAuthService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	t.Run("正常注册", func(t *testing.T) {
		// 模拟邮箱不存在
		mockRepo.On("GetByEmail", "new@example.com").Return(nil, nil).Once()
		mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(nil).Once()

		user, err := authService.Register("new@example.com", "password123")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "new@example.com", user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("邮箱已存在", func(t *testing.T) {
		existingUser := &model.User{Email: "existing@example.com"}
		mockRepo.On("GetByEmail", "existing@example.com").Return(existingUser, nil).Once()

		user, err := authService.Register("existing@example.com", "password123")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, bizerrors.ErrEmailExists, err)
		mockRepo.AssertExpectations(t)
	})
}

// TestAuthService_Login 测试用户登录
func TestAuthService_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	// 生成有效的密码哈希
	passwordHash, _ := utils.HashPassword("password123")

	t.Run("正常登录", func(t *testing.T) {
		// 创建一个已存在的用户
		user := &model.User{
			Email:        "test@example.com",
			PasswordHash: passwordHash,
			Status:       1,
			Role:         "user",
		}
		mockRepo.On("GetByEmail", "test@example.com").Return(user, nil).Once()

		tokenPair, loginUser, err := authService.Login("test@example.com", "password123")
		assert.NoError(t, err)
		assert.NotNil(t, tokenPair)
		assert.NotNil(t, loginUser)
		assert.Equal(t, "test@example.com", loginUser.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("用户不存在", func(t *testing.T) {
		mockRepo.On("GetByEmail", "nonexistent@example.com").Return(nil, nil).Once()

		tokenPair, user, err := authService.Login("nonexistent@example.com", "password123")
		assert.Error(t, err)
		assert.Nil(t, tokenPair)
		assert.Nil(t, user)
		assert.Equal(t, bizerrors.ErrInvalidCredentials, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("密码错误", func(t *testing.T) {
		user := &model.User{
			Email:        "test@example.com",
			PasswordHash: passwordHash,
			Status:       1,
			Role:         "user",
		}
		mockRepo.On("GetByEmail", "test@example.com").Return(user, nil).Once()

		tokenPair, loginUser, err := authService.Login("test@example.com", "wrongpassword")
		assert.Error(t, err)
		assert.Nil(t, tokenPair)
		assert.Nil(t, loginUser)
		assert.Equal(t, bizerrors.ErrInvalidCredentials, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("账号被封禁", func(t *testing.T) {
		user := &model.User{
			Email:        "banned@example.com",
			PasswordHash: passwordHash,
			Status:       0, // 被封禁
			Role:         "user",
		}
		mockRepo.On("GetByEmail", "banned@example.com").Return(user, nil).Once()

		tokenPair, loginUser, err := authService.Login("banned@example.com", "password123")
		assert.Error(t, err)
		assert.Nil(t, tokenPair)
		assert.Nil(t, loginUser)
		assert.Equal(t, bizerrors.ErrAccountBanned, err)
		mockRepo.AssertExpectations(t)
	})
}

// TestUserService_GetUserByID 测试获取用户
func TestUserService_GetUserByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)

	t.Run("用户存在", func(t *testing.T) {
		user := &model.User{
			ID:    1,
			Email: "test@example.com",
		}
		mockRepo.On("GetByID", uint(1)).Return(user, nil).Once()

		result, err := userService.GetUserByID(1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("用户不存在", func(t *testing.T) {
		mockRepo.On("GetByID", uint(999)).Return(nil, nil).Once()

		result, err := userService.GetUserByID(999)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, bizerrors.ErrUserNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

// TestUserService_ChangePassword 测试修改密码
func TestUserService_ChangePassword(t *testing.T) {
	// 生成有效的密码哈希
	passwordHash, _ := utils.HashPassword("password123")

	t.Run("正常修改密码", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		userService := NewUserService(mockRepo)

		user := &model.User{
			ID:           1,
			Email:        "test@example.com",
			PasswordHash: passwordHash,
		}
		mockRepo.On("GetByID", uint(1)).Return(user, nil).Once()
		mockRepo.On("Update", mock.AnythingOfType("*model.User")).Return(nil).Once()

		err := userService.ChangePassword(1, "password123", "newpassword123")
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("旧密码错误", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		userService := NewUserService(mockRepo)

		user := &model.User{
			ID:           1,
			Email:        "test@example.com",
			PasswordHash: passwordHash,
		}
		mockRepo.On("GetByID", uint(1)).Return(user, nil).Once()

		err := userService.ChangePassword(1, "wrongpassword", "newpassword123")
		assert.Error(t, err)
		assert.Equal(t, bizerrors.ErrInvalidPassword, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("用户不存在", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		userService := NewUserService(mockRepo)

		mockRepo.On("GetByID", uint(999)).Return(nil, nil).Once()

		err := userService.ChangePassword(999, "password123", "newpassword123")
		assert.Error(t, err)
		assert.Equal(t, bizerrors.ErrUserNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}
