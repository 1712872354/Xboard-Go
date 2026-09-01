package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// PluginHandler 插件处理器
type PluginHandler struct {
	pluginService service.PluginService
}

// NewPluginHandler 创建插件处理器
func NewPluginHandler(pluginService service.PluginService) *PluginHandler {
	return &PluginHandler{
		pluginService: pluginService,
	}
}

// CreatePluginRequest 创建插件请求
type CreatePluginRequest struct {
	Name        string `json:"name" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Homepage    string `json:"homepage"`
	Config      string `json:"config"`
	Type        string `json:"type"`
}

// CreatePlugin 创建插件（管理员）
func (h *PluginHandler) CreatePlugin(c *gin.Context) {
	var req CreatePluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	plugin, err := h.pluginService.Create(
		req.Name, req.Title, req.Description,
		req.Version, req.Author, req.Homepage, req.Config, req.Type,
	)
	if err != nil {
		response.InternalError(c, "创建插件失败："+err.Error())
		return
	}

	response.Success(c, plugin)
}

// GetPlugin 获取插件详情（管理员）
func (h *PluginHandler) GetPlugin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的插件ID")
		return
	}

	plugin, err := h.pluginService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "插件不存在")
		return
	}

	response.Success(c, plugin)
}

// UpdatePluginRequest 更新插件请求
type UpdatePluginRequest struct {
	Name        string `json:"name" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Homepage    string `json:"homepage"`
	Config      string `json:"config"`
	Status      int    `json:"status"`
}

// UpdatePlugin 更新插件（管理员）
func (h *PluginHandler) UpdatePlugin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的插件ID")
		return
	}

	var req UpdatePluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	plugin, err := h.pluginService.Update(
		uint(id), req.Name, req.Title, req.Description,
		req.Version, req.Author, req.Homepage, req.Config, req.Status,
	)
	if err != nil {
		response.InternalError(c, "更新插件失败："+err.Error())
		return
	}

	response.Success(c, plugin)
}

// DeletePlugin 删除插件（管理员）
func (h *PluginHandler) DeletePlugin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的插件ID")
		return
	}

	// 检查是否为受保护插件
	plugin, err := h.pluginService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "插件不存在")
		return
	}
	if plugin.IsProtected {
		response.BadRequest(c, "系统内置插件不可删除")
		return
	}

	if err := h.pluginService.Delete(uint(id)); err != nil {
		response.InternalError(c, "删除插件失败")
		return
	}

	response.Success(c, nil)
}

// UploadPlugin 上传插件（管理员）
func (h *PluginHandler) UploadPlugin(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传插件文件")
		return
	}

	// 验证文件扩展名
	if len(file.Filename) < 4 || file.Filename[len(file.Filename)-4:] != ".zip" {
		response.BadRequest(c, "仅支持 .zip 格式的插件包")
		return
	}

	// 暂时只返回成功提示（实际解压逻辑后续实现）
	// 调用 service.Create 创建插件记录，基本信息从文件名提取
	plugin, err := h.pluginService.Create(
		file.Filename[:len(file.Filename)-4], // name: 去掉 .zip 后缀
		file.Filename[:len(file.Filename)-4], // title: 同 name
		"",                                    // description
		"1.0.0",                               // version
		"",                                    // author
		"",                                    // homepage
		"",                                    // config
		"feature",                             // type
	)
	if err != nil {
		response.InternalError(c, "创建插件记录失败："+err.Error())
		return
	}

	response.Success(c, plugin)
}

// ListPlugins 插件列表（管理员）
func (h *PluginHandler) ListPlugins(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	plugins, total, err := h.pluginService.List(page, pageSize)
	if err != nil {
		response.InternalError(c, "获取插件列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  plugins,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// UpdatePluginStatusRequest 更新插件状态请求
type UpdatePluginStatusRequest struct {
	Status int `json:"status"`
}

// UpdatePluginStatus 更新插件状态（管理员）
func (h *PluginHandler) UpdatePluginStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的插件ID")
		return
	}

	var req UpdatePluginStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.pluginService.UpdateStatus(uint(id), req.Status); err != nil {
		response.InternalError(c, "更新插件状态失败："+err.Error())
		return
	}

	response.Success(c, nil)
}

// EnablePlugin 启用插件（管理员）
func (h *PluginHandler) EnablePlugin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的插件ID")
		return
	}

	if err := h.pluginService.Enable(uint(id)); err != nil {
		response.InternalError(c, "启用插件失败："+err.Error())
		return
	}

	response.Success(c, nil)
}

// DisablePlugin 禁用插件（管理员）
func (h *PluginHandler) DisablePlugin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的插件ID")
		return
	}

	if err := h.pluginService.Disable(uint(id)); err != nil {
		response.InternalError(c, "禁用插件失败："+err.Error())
		return
	}

	response.Success(c, nil)
}
