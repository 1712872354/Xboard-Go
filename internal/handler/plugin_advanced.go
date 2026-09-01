package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"
)

// PluginAdvancedHandler 插件高级处理器
type PluginAdvancedHandler struct {
	pluginDir string
}

// NewPluginAdvancedHandler 创建插件高级处理器
func NewPluginAdvancedHandler(pluginDir string) *PluginAdvancedHandler {
	return &PluginAdvancedHandler{
		pluginDir: pluginDir,
	}
}

// GetPluginTypes 获取插件类型列表
// @Summary 插件类型
// @Description 获取所有插件类型
// @Tags 管理员-插件
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/plugins/types [get]
func (h *PluginAdvancedHandler) GetPluginTypes(c *gin.Context) {
	types := []gin.H{
		{"id": "payment", "name": "支付插件"},
		{"id": "notification", "name": "通知插件"},
		{"id": "auth", "name": "认证插件"},
		{"id": "theme", "name": "主题插件"},
		{"id": "tool", "name": "工具插件"},
	}

	response.Success(c, types)
}

// InstallPlugin 安装插件
// @Summary 安装插件
// @Description 安装指定插件
// @Tags 管理员-插件
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "插件ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/plugins/{id}/install [post]
func (h *PluginAdvancedHandler) InstallPlugin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的插件ID")
		return
	}

	db := database.Get()
	var plugin model.Plugin
	if err := db.First(&plugin, id).Error; err != nil {
		response.NotFound(c, "插件不存在")
		return
	}

	// 检查插件目录是否存在
	pluginPath := filepath.Join(h.pluginDir, plugin.Name)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		response.BadRequest(c, "插件文件不存在")
		return
	}

	// 更新插件状态
	if err := db.Model(&plugin).Updates(map[string]interface{}{
		"status":     1,
		"installed":  true,
	}).Error; err != nil {
		response.InternalError(c, "安装插件失败")
		return
	}

	response.Success(c, gin.H{
		"message": "插件安装成功",
	})
}

// UninstallPlugin 卸载插件
// @Summary 卸载插件
// @Description 卸载指定插件
// @Tags 管理员-插件
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "插件ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/plugins/{id}/uninstall [post]
func (h *PluginAdvancedHandler) UninstallPlugin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的插件ID")
		return
	}

	db := database.Get()
	var plugin model.Plugin
	if err := db.First(&plugin, id).Error; err != nil {
		response.NotFound(c, "插件不存在")
		return
	}

	// 更新插件状态
	if err := db.Model(&plugin).Updates(map[string]interface{}{
		"status":     0,
		"installed":  false,
	}).Error; err != nil {
		response.InternalError(c, "卸载插件失败")
		return
	}

	response.Success(c, gin.H{
		"message": "插件卸载成功",
	})
}

// GetPluginConfig 获取插件配置
// @Summary 获取插件配置
// @Description 获取指定插件的配置
// @Tags 管理员-插件
// @Produce json
// @Security Bearer
// @Param id path int true "插件ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/plugins/{id}/config [get]
func (h *PluginAdvancedHandler) GetPluginConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的插件ID")
		return
	}

	db := database.Get()
	var plugin model.Plugin
	if err := db.First(&plugin, id).Error; err != nil {
		response.NotFound(c, "插件不存在")
		return
	}

	// 解析配置
	var config map[string]interface{}
	if plugin.Config != "" {
		json.Unmarshal([]byte(plugin.Config), &config)
	}

	if config == nil {
		config = map[string]interface{}{}
	}

	response.Success(c, config)
}

// UpdatePluginConfigRequest 更新插件配置请求
type UpdatePluginConfigRequest struct {
	Config map[string]interface{} `json:"config" binding:"required"`
}

// UpdatePluginConfig 更新插件配置
// @Summary 更新插件配置
// @Description 更新指定插件的配置
// @Tags 管理员-插件
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "插件ID"
// @Param request body UpdatePluginConfigRequest true "插件配置"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/plugins/{id}/config [put]
func (h *PluginAdvancedHandler) UpdatePluginConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的插件ID")
		return
	}

	var req UpdatePluginConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	db := database.Get()
	var plugin model.Plugin
	if err := db.First(&plugin, id).Error; err != nil {
		response.NotFound(c, "插件不存在")
		return
	}

	// 序列化配置
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		response.InternalError(c, "配置格式错误")
		return
	}

	// 更新配置
	if err := db.Model(&plugin).Update("config", string(configJSON)).Error; err != nil {
		response.InternalError(c, "更新配置失败")
		return
	}

	response.Success(c, nil)
}

// UpgradePlugin 升级插件
// @Summary 升级插件
// @Description 升级指定插件到新版本
// @Tags 管理员-插件
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "插件ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/plugins/{id}/upgrade [post]
func (h *PluginAdvancedHandler) UpgradePlugin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的插件ID")
		return
	}

	db := database.Get()
	var plugin model.Plugin
	if err := db.First(&plugin, id).Error; err != nil {
		response.NotFound(c, "插件不存在")
		return
	}

	// 检查插件目录
	pluginPath := filepath.Join(h.pluginDir, plugin.Name)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		response.BadRequest(c, "插件文件不存在")
		return
	}

	// 读取插件配置文件获取新版本信息
	configPath := filepath.Join(pluginPath, "plugin.json")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		response.InternalError(c, "读取插件配置失败")
		return
	}

	var pluginConfig struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(configData, &pluginConfig); err != nil {
		response.InternalError(c, "解析插件配置失败")
		return
	}

	// 更新版本
	if err := db.Model(&plugin).Update("version", pluginConfig.Version).Error; err != nil {
		response.InternalError(c, "升级插件失败")
		return
	}

	response.Success(c, gin.H{
		"message":    "插件升级成功",
		"old_version": plugin.Version,
		"new_version": pluginConfig.Version,
	})
}

// GetPluginConfigTemplate 获取插件配置模板
// @Summary 获取配置模板
// @Description 获取插件的配置模板
// @Tags 管理员-插件
// @Produce json
// @Security Bearer
// @Param id path int true "插件ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/plugins/{id}/config-template [get]
func (h *PluginAdvancedHandler) GetPluginConfigTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的插件ID")
		return
	}

	db := database.Get()
	var plugin model.Plugin
	if err := db.First(&plugin, id).Error; err != nil {
		response.NotFound(c, "插件不存在")
		return
	}

	// 读取插件配置文件
	pluginPath := filepath.Join(h.pluginDir, plugin.Name)
	configPath := filepath.Join(pluginPath, "config.template.json")
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 如果没有配置模板，返回空对象
		response.Success(c, map[string]interface{}{})
		return
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		response.InternalError(c, "读取配置模板失败")
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		response.InternalError(c, "解析配置模板失败")
		return
	}

	response.Success(c, config)
}

// ReloadPlugins 重新加载插件
// @Summary 重新加载插件
// @Description 重新加载所有插件
// @Tags 管理员-插件
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/plugins/reload [post]
func (h *PluginAdvancedHandler) ReloadPlugins(c *gin.Context) {
	// 扫描插件目录
	entries, err := os.ReadDir(h.pluginDir)
	if err != nil {
		response.InternalError(c, "读取插件目录失败")
		return
	}

	db := database.Get()
	var loadedCount int

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 读取插件配置
		configPath := filepath.Join(h.pluginDir, entry.Name(), "plugin.json")
		configData, err := os.ReadFile(configPath)
		if err != nil {
			continue
		}

		var pluginConfig struct {
			Name        string `json:"name"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Version     string `json:"version"`
			Author      string `json:"author"`
			Type        string `json:"type"`
		}
		if err := json.Unmarshal(configData, &pluginConfig); err != nil {
			continue
		}

		// 检查插件是否已存在
		var existingPlugin model.Plugin
		if err := db.Where("name = ?", pluginConfig.Name).First(&existingPlugin).Error; err != nil {
			// 创建新插件记录
			newPlugin := model.Plugin{
				Name:        pluginConfig.Name,
				Title:       pluginConfig.Title,
				Description: pluginConfig.Description,
				Version:     pluginConfig.Version,
				Author:      pluginConfig.Author,
				Type:        pluginConfig.Type,
				Status:      0,
			}
			db.Create(&newPlugin)
		} else {
			// 更新现有插件版本
			db.Model(&existingPlugin).Update("version", pluginConfig.Version)
		}

		loadedCount++
	}

	response.Success(c, gin.H{
		"message": fmt.Sprintf("已加载 %d 个插件", loadedCount),
		"count":   loadedCount,
	})
}
