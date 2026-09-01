package handler

import (
	"xboard-go/internal/middleware"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// SessionHandler 会话管理处理器
type SessionHandler struct {
	sessionService service.SessionService
}

// NewSessionHandler 创建会话管理处理器
func NewSessionHandler(sessionService service.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// GetActiveSessions 获取活跃会话列表
// @Summary 获取活跃会话
// @Description 获取当前用户的所有活跃会话
// @Tags 用户-会话
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=[]service.SessionInfo}
// @Router /api/v1/user/sessions [get]
func (h *SessionHandler) GetActiveSessions(c *gin.Context) {
	userID := middleware.GetUserID(c)

	sessions, err := h.sessionService.GetActiveSessions(userID)
	if err != nil {
		response.InternalError(c, "获取会话列表失败")
		return
	}

	response.Success(c, sessions)
}

// RemoveSessionRequest 移除会话请求
type RemoveSessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// RemoveSession 移除指定会话
// @Summary 移除会话
// @Description 移除当前用户的指定会话
// @Tags 用户-会话
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body RemoveSessionRequest true "会话信息"
// @Success 200 {object} response.Response
// @Router /api/v1/user/sessions/remove [post]
func (h *SessionHandler) RemoveSession(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req RemoveSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if err := h.sessionService.RemoveSession(userID, req.SessionID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// CheckLogin 检查登录状态
// @Summary 检查登录状态
// @Description 检查当前用户是否已登录
// @Tags 用户-认证
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/user/check-login [get]
func (h *SessionHandler) CheckLogin(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Success(c, gin.H{
			"is_login": false,
		})
		return
	}

	// 检查是否为管理员
	role, _ := c.Get("user_role")
	isAdmin := role == "admin"

	response.Success(c, gin.H{
		"is_login":  true,
		"user_id":   userID,
		"is_admin":  isAdmin,
	})
}
