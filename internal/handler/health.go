package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"xboard-go/config"
	"xboard-go/pkg/response"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck 健康检查响应
type HealthCheck struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
	AppName   string `json:"app_name"`
}

// Check 健康检查接口
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags 系统
// @Produce json
// @Success 200 {object} response.Response{data=HealthCheck}
// @Router /healthz [get]
func (h *HealthHandler) Check(c *gin.Context) {
	cfg := config.Get()
	response.Success(c, HealthCheck{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   cfg.App.Version,
		AppName:   cfg.App.Name,
	})
}
