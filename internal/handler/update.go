package handler

import (
	"xboard-go/internal/service"
	"xboard-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// UpdateHandler 系统更新处理器
type UpdateHandler struct {
	updateService service.UpdateService
}

// NewUpdateHandler 创建更新处理器
func NewUpdateHandler(version string) *UpdateHandler {
	return &UpdateHandler{
		updateService: service.NewUpdateService(version),
	}
}

// CheckUpdate 检查更新
func (h *UpdateHandler) CheckUpdate(c *gin.Context) {
	info, err := h.updateService.CheckUpdate()
	if err != nil {
		response.Fail(c, 500, "检查更新失败: "+err.Error())
		return
	}

	response.Success(c, info)
}

// ExecuteUpdate 执行更新
func (h *UpdateHandler) ExecuteUpdate(c *gin.Context) {
	if err := h.updateService.ExecuteUpdate(); err != nil {
		response.Fail(c, 500, "更新失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "更新完成，服务将在稍后自动重启",
	})
}
