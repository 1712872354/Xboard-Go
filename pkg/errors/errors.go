package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode 业务错误码类型
type ErrorCode int

// 通用错误码 (1xxx)
const (
	ErrCodeInternal     ErrorCode = 1000
	ErrCodeInvalidParam ErrorCode = 1001
	ErrCodeNotFound     ErrorCode = 1002
	ErrCodeConflict     ErrorCode = 1003
	ErrCodeUnauthorized ErrorCode = 1004
	ErrCodeForbidden    ErrorCode = 1005
)

// 用户相关错误码 (2xxx)
const (
	ErrCodeUserNotFound      ErrorCode = 2001
	ErrCodeEmailExists       ErrorCode = 2002
	ErrCodeInvalidPassword   ErrorCode = 2003
	ErrCodeAccountBanned     ErrorCode = 2004
	ErrCodeInvalidCredentials ErrorCode = 2005
)

// 订单相关错误码 (3xxx)
const (
	ErrCodeOrderNotFound   ErrorCode = 3001
	ErrCodeOrderPaid       ErrorCode = 3002
	ErrCodeOrderCancelled  ErrorCode = 3003
	ErrCodePaymentFailed   ErrorCode = 3004
)

// 套餐相关错误码 (4xxx)
const (
	ErrCodePlanNotFound ErrorCode = 4001
	ErrCodePlanInactive ErrorCode = 4002
)

// 节点相关错误码 (5xxx)
const (
	ErrCodeNodeNotFound ErrorCode = 5001
	ErrCodeNodeOffline  ErrorCode = 5002
)

// BusinessError 业务错误类型
type BusinessError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Err        error     `json:"-"`
	HTTPStatus int       `json:"-"`
}

// Error 实现error接口
func (e *BusinessError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 实现errors.Unwrap接口
func (e *BusinessError) Unwrap() error {
	return e.Err
}

// GetHTTPStatus 获取HTTP状态码
func (e *BusinessError) GetHTTPStatus() int {
	if e.HTTPStatus > 0 {
		return e.HTTPStatus
	}
	return http.StatusBadRequest
}

// NewBusinessError 创建业务错误
func NewBusinessError(code ErrorCode, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// WrapError 包装底层错误
func WrapError(code ErrorCode, message string, err error) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WithHTTPStatus 设置HTTP状态码
func (e *BusinessError) WithHTTPStatus(status int) *BusinessError {
	e.HTTPStatus = status
	return e
}

// 预定义的业务错误
var (
	ErrInternal = NewBusinessError(ErrCodeInternal, "服务器内部错误").
			WithHTTPStatus(http.StatusInternalServerError)

	ErrInvalidParam = NewBusinessError(ErrCodeInvalidParam, "参数错误")

	ErrNotFound = NewBusinessError(ErrCodeNotFound, "资源不存在").
			WithHTTPStatus(http.StatusNotFound)

	ErrConflict = NewBusinessError(ErrCodeConflict, "资源冲突")

	ErrUnauthorized = NewBusinessError(ErrCodeUnauthorized, "未授权").
			WithHTTPStatus(http.StatusUnauthorized)

	ErrForbidden = NewBusinessError(ErrCodeForbidden, "禁止访问").
			WithHTTPStatus(http.StatusForbidden)

	// 用户相关错误
	ErrUserNotFound = NewBusinessError(ErrCodeUserNotFound, "用户不存在").
			WithHTTPStatus(http.StatusNotFound)

	ErrEmailExists = NewBusinessError(ErrCodeEmailExists, "邮箱已被注册")

	ErrInvalidPassword = NewBusinessError(ErrCodeInvalidPassword, "密码错误")

	ErrAccountBanned = NewBusinessError(ErrCodeAccountBanned, "账号已被封禁").
			WithHTTPStatus(http.StatusForbidden)

	ErrInvalidCredentials = NewBusinessError(ErrCodeInvalidCredentials, "邮箱或密码错误")

	// 订单相关错误
	ErrOrderNotFound = NewBusinessError(ErrCodeOrderNotFound, "订单不存在").
				WithHTTPStatus(http.StatusNotFound)

	ErrOrderPaid = NewBusinessError(ErrCodeOrderPaid, "订单已支付")

	ErrOrderCancelled = NewBusinessError(ErrCodeOrderCancelled, "订单已取消")

	ErrPaymentFailed = NewBusinessError(ErrCodePaymentFailed, "支付失败")

	// 套餐相关错误
	ErrPlanNotFound = NewBusinessError(ErrCodePlanNotFound, "套餐不存在").
			WithHTTPStatus(http.StatusNotFound)

	ErrPlanInactive = NewBusinessError(ErrCodePlanInactive, "套餐已下架")

	// 节点相关错误
	ErrNodeNotFound = NewBusinessError(ErrCodeNodeNotFound, "节点不存在").
			WithHTTPStatus(http.StatusNotFound)

	ErrNodeOffline = NewBusinessError(ErrCodeNodeOffline, "节点离线")
)

// IsBusinessError 检查是否为业务错误
func IsBusinessError(err error) bool {
	_, ok := err.(*BusinessError)
	return ok
}

// GetBusinessError 获取业务错误，如果不是业务错误则返回nil
func GetBusinessError(err error) *BusinessError {
	if bizErr, ok := err.(*BusinessError); ok {
		return bizErr
	}
	return nil
}
