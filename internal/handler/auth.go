package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"xboard-go/internal/middleware"
	"xboard-go/internal/model"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// NotifyUserChange is called after a user is updated/deleted via the admin API.
// Set by cmd/server/main.go to bridge HTTP handlers to the gRPC broadcaster.
var NotifyUserChange func()

// AuthHandler 认证处理器
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=50"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest 刷新token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	User         *model.User `json:"user"`
}

// Register 用户注册
// @Summary 用户注册
// @Description 用户注册新账号
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 200 {object} response.Response{data=model.User}
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	user, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Notify connected gRPC nodes about the new user.
	if NotifyUserChange != nil {
		go NotifyUserChange()
	}

	response.Success(c, user)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=LoginResponse}
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	tokenPair, user, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         user,
	})
}

// RefreshToken 刷新token
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌"
// @Success 200 {object} response.Response{data=jwt.TokenPair}
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters")
		return
	}

	tokenPair, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, tokenPair)
}

// UserHandler 用户处理器
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ProfileResponse 用户信息响应
type ProfileResponse struct {
	ID               uint    `json:"id"`
	Email            string  `json:"email"`
	Role             string  `json:"role"`
	Status           int     `json:"status"`
	TrafficLimit     int64   `json:"traffic_limit"`
	UsedTraffic      int64   `json:"used_traffic"`
	ExpiredAt        string  `json:"expired_at"`
	SubscribeToken   string  `json:"subscribe_token"`
	Balance          float64 `json:"balance"`
	Commission       float64 `json:"commission"`
	PlanID           *uint   `json:"plan_id"`
	TwoFactorEnabled bool    `json:"two_factor_enabled"`
	OnlineCount      int     `json:"online_count"`
	RemindExpire     bool    `json:"remind_expire"`
	RemindTraffic    bool    `json:"remind_traffic"`
	CreatedAt        string  `json:"created_at"`
}

// GetProfile 获取当前用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的信息
// @Tags 用户
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=ProfileResponse}
// @Router /api/v1/user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	expiredAt := ""
	if user.ExpiredAt != nil {
		expiredAt = user.ExpiredAt.Format("2006-01-02 15:04:05")
	}

	profile := ProfileResponse{
		ID:               user.ID,
		Email:            user.Email,
		Role:             user.Role,
		Status:           user.Status,
		TrafficLimit:     user.TrafficLimit,
		UsedTraffic:      user.UsedTraffic,
		ExpiredAt:        expiredAt,
		SubscribeToken:   user.SubscribeToken,
		Balance:          user.Balance,
		Commission:       user.Commission,
		PlanID:           user.PlanID,
		TwoFactorEnabled: user.TwoFactorEnabled,
		OnlineCount:      user.OnlineCount,
		RemindExpire:     user.RemindExpire,
		RemindTraffic:    user.RemindTraffic,
		CreatedAt:        user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	response.Success(c, profile)
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Email string `json:"email" binding:"omitempty,email"`
}

// UpdateProfile 更新用户资料
// @Summary 更新用户资料
// @Description 更新当前用户的资料
// @Tags 用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body UpdateProfileRequest true "更新信息"
// @Success 200 {object} response.Response{data=model.User}
// @Router /api/v1/user/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	user, err := h.userService.UpdateProfile(userID, req.Email)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, user)
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=50"`
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户的密码
// @Tags 用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body ChangePasswordRequest true "密码信息"
// @Success 200 {object} response.Response
// @Router /api/v1/user/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ResetSubscribeToken 重置订阅token
// @Summary 重置订阅Token
// @Description 重置当前用户的订阅链接token
// @Tags 用户
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /api/v1/user/subscribe/reset [post]
func (h *UserHandler) ResetSubscribeToken(c *gin.Context) {
	userID := middleware.GetUserID(c)

	newToken, err := h.userService.ResetSubscribeToken(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// Token changed, notify nodes to refresh user list.
	if NotifyUserChange != nil {
		go NotifyUserChange()
	}

	response.Success(c, gin.H{"subscribe_token": newToken})
}

// ListUsers 用户列表（管理员）
// @Summary 用户列表
// @Description 获取用户列表（管理员权限）
// @Tags 管理员-用户
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param keyword query string false "搜索关键词（邮箱）"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page := c.GetInt("page")
	pageSize := c.GetInt("page_size")
	keyword := c.Query("keyword")

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	users, total, err := h.userService.ListUsers(page, pageSize, keyword)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	Status int `json:"status" binding:"required,oneof=0 1"`
}

// UpdateUserStatus 更新用户状态
// @Summary 更新用户状态
// @Description 启用或禁用用户（管理员权限）
// @Tags 管理员-用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Param request body UpdateUserStatusRequest true "状态"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/users/{id}/status [put]
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	var userID struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&userID); err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters")
		return
	}

	if err := h.userService.UpdateUserStatus(userID.ID, req.Status); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if NotifyUserChange != nil {
		go NotifyUserChange()
	}

	response.Success(c, nil)
}

// UpdateUserRoleRequest 更新用户角色请求
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user admin"`
}

// UpdateUserRole 更新用户角色
// @Summary 更新用户角色
// @Description 修改用户角色（管理员权限）
// @Tags 管理员-用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Param request body UpdateUserRoleRequest true "角色"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/users/{id}/role [put]
func (h *UserHandler) UpdateUserRole(c *gin.Context) {
	var userID struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&userID); err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters")
		return
	}

	if err := h.userService.UpdateUserRole(userID.ID, req.Role); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if NotifyUserChange != nil {
		go NotifyUserChange()
	}

	response.Success(c, nil)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除用户（管理员权限）
// @Tags 管理员-用户
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	var userID struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&userID); err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.userService.DeleteUser(userID.ID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if NotifyUserChange != nil {
		go NotifyUserChange()
	}

	response.Success(c, nil)
}

// GetUser 获取单个用户详情（管理员）
// @Summary 获取用户详情
// @Description 获取指定用户的详细信息（管理员权限）
// @Tags 管理员-用户
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=model.User}
// @Router /api/v1/admin/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	var userID struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&userID); err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUserByID(userID.ID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Code:    0,
		Message: "success",
		Data:    user,
	})
}

// AdminUpdateUserRequest 管理员更新用户请求
type AdminUpdateUserRequest struct {
	Email        string   `json:"email"`
	Password     string   `json:"password"`
	TrafficLimit *int64   `json:"traffic_limit"`
	ExpiredAt    *string  `json:"expired_at"`
	PlanID       *uint    `json:"plan_id"`
	Balance      *float64 `json:"balance"`
	Commission   *float64 `json:"commission"`
}

// AdminUpdateUser 管理员更新用户信息
// @Summary 管理员更新用户
// @Description 管理员更新用户详细信息（管理员权限）
// @Tags 管理员-用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Param request body AdminUpdateUserRequest true "更新信息"
// @Success 200 {object} response.Response{data=model.User}
// @Router /api/v1/admin/users/{id} [put]
func (h *UserHandler) AdminUpdateUser(c *gin.Context) {
	var userID struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&userID); err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	user, err := h.userService.AdminUpdateUser(
		userID.ID,
		req.Email,
		req.Password,
		req.TrafficLimit,
		req.ExpiredAt,
		req.PlanID,
		req.Balance,
		req.Commission,
	)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if NotifyUserChange != nil {
		go NotifyUserChange()
	}

	response.Success(c, user)
}

// ForgetPasswordRequest 忘记密码请求
type ForgetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgetPassword 忘记密码
// @Summary 忘记密码
// @Description 发送重置密码邮件
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body ForgetPasswordRequest true "邮箱"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/forget [post]
func (h *AuthHandler) ForgetPassword(c *gin.Context) {
	var req ForgetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if err := h.authService.ForgetPassword(req.Email); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=50"`
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 使用token重置密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "重置信息"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/reset [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// SendVerificationCodeRequest 发送验证码请求
type SendVerificationCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// SendVerificationCode 发送验证码
// @Summary 发送验证码
// @Description 发送邮箱验证码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body SendVerificationCodeRequest true "邮箱"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/send-code [post]
func (h *AuthHandler) SendVerificationCode(c *gin.Context) {
	var req SendVerificationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if err := h.authService.SendVerificationCode(req.Email); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// VerifyEmailRequest 验证邮箱请求
type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

// VerifyEmail 验证邮箱
// @Summary 验证邮箱
// @Description 验证邮箱验证码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "验证信息"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/verify [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if err := h.authService.VerifyEmail(req.Email, req.Code); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GenerateUsersRequest 批量生成用户请求
type GenerateUsersRequest struct {
	Count     int    `json:"count" binding:"required,min=1,max=1000"`
	Prefix    string `json:"prefix"`
	Password  string `json:"password"`
	PlanID    int    `json:"plan_id"`
	ExpiredAt string `json:"expired_at"`
}

// GenerateUsers 批量生成用户（管理员）
// @Summary 批量生成用户
// @Description 批量生成用户账号（管理员权限）
// @Tags 管理员-用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body GenerateUsersRequest true "生成参数"
// @Success 200 {object} response.Response{data=[]model.User}
// @Router /api/v1/admin/users/generate [post]
func (h *UserHandler) GenerateUsers(c *gin.Context) {
	var req GenerateUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	users, err := h.userService.GenerateUsers(req.Count, req.Prefix, req.Password, req.PlanID, req.ExpiredAt)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if NotifyUserChange != nil {
		go NotifyUserChange()
	}

	response.Success(c, gin.H{
		"users": users,
		"count": len(users),
	})
}

// ExportUsersCSV 导出用户CSV（管理员）
// @Summary 导出用户CSV
// @Description 导出用户列表为CSV格式（管理员权限）
// @Tags 管理员-用户
// @Produce text/csv
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(100)
// @Param keyword query string false "搜索关键词（邮箱）"
// @Success 200 {string} string "CSV文件内容"
// @Router /api/v1/admin/users/export [get]
func (h *UserHandler) ExportUsersCSV(c *gin.Context) {
	page := 1
	pageSize := 100
	keyword := c.Query("keyword")

	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(c.Query("page_size")); err == nil && ps > 0 && ps <= 1000 {
		pageSize = ps
	}

	csvContent, err := h.userService.ExportUsersCSV(page, pageSize, keyword)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	filename := fmt.Sprintf("users_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.String(http.StatusOK, csvContent)
}
