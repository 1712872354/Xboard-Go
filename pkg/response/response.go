package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 统一响应状态码
const (
	CodeSuccess         = 0
	CodeBadRequest      = 400
	CodeUnauthorized    = 401
	CodeForbidden       = 403
	CodeNotFound        = 404
	CodeTooManyRequests = 429
	CodeInternalError   = 500
)

// Response 统一响应结构体
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest 参数错误
func BadRequest(c *gin.Context, message string) {
	Fail(c, CodeBadRequest, message)
}

// Unauthorized 未认证
func Unauthorized(c *gin.Context, message string) {
	Fail(c, CodeUnauthorized, message)
}

// Forbidden 无权限
func Forbidden(c *gin.Context, message string) {
	Fail(c, CodeForbidden, message)
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	Fail(c, CodeNotFound, message)
}

// InternalError 服务器内部错误
func InternalError(c *gin.Context, message string) {
	Fail(c, CodeInternalError, message)
}

// TooManyRequests 请求过于频繁（限流）
// 使用 HTTP 429 状态码，并附带 Retry-After 头
func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, Response{
		Code:    CodeTooManyRequests,
		Message: message,
	})
}
