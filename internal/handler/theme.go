package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"
)

// ThemeHandler 主题处理器
type ThemeHandler struct {
	themeDir string
}

// NewThemeHandler 创建主题处理器
func NewThemeHandler(themeDir string) *ThemeHandler {
	return &ThemeHandler{
		themeDir: themeDir,
	}
}

// ListThemes 获取主题列表
// @Summary 主题列表
// @Description 获取所有主题列表
// @Tags 管理员-主题
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/themes [get]
func (h *ThemeHandler) ListThemes(c *gin.Context) {
	db := database.Get()

	var themes []model.Theme
	if err := db.Order("is_default DESC, name ASC").Find(&themes).Error; err != nil {
		response.InternalError(c, "获取主题列表失败")
		return
	}

	if themes == nil {
		themes = []model.Theme{}
	}

	response.Success(c, themes)
}

// GetTheme 获取主题详情
// @Summary 主题详情
// @Description 获取指定主题的详细信息
// @Tags 管理员-主题
// @Produce json
// @Security Bearer
// @Param id path int true "主题ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/themes/{id} [get]
func (h *ThemeHandler) GetTheme(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的主题ID")
		return
	}

	db := database.Get()
	var theme model.Theme
	if err := db.First(&theme, id).Error; err != nil {
		response.NotFound(c, "主题不存在")
		return
	}

	response.Success(c, theme)
}

// GetThemeConfig 获取主题配置
// @Summary 获取主题配置
// @Description 获取指定主题的配置信息
// @Tags 管理员-主题
// @Produce json
// @Security Bearer
// @Param id path int true "主题ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/themes/{id}/config [get]
func (h *ThemeHandler) GetThemeConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的主题ID")
		return
	}

	db := database.Get()
	var theme model.Theme
	if err := db.First(&theme, id).Error; err != nil {
		response.NotFound(c, "主题不存在")
		return
	}

	// 解析配置
	var config map[string]interface{}
	if theme.Config != "" {
		json.Unmarshal([]byte(theme.Config), &config)
	}

	if config == nil {
		config = map[string]interface{}{}
	}

	response.Success(c, config)
}

// UpdateThemeConfigRequest 更新主题配置请求
type UpdateThemeConfigRequest struct {
	Config map[string]interface{} `json:"config" binding:"required"`
}

// UpdateThemeConfig 更新主题配置
// @Summary 更新主题配置
// @Description 更新指定主题的配置
// @Tags 管理员-主题
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "主题ID"
// @Param request body UpdateThemeConfigRequest true "主题配置"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/themes/{id}/config [put]
func (h *ThemeHandler) UpdateThemeConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的主题ID")
		return
	}

	var req UpdateThemeConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	db := database.Get()
	var theme model.Theme
	if err := db.First(&theme, id).Error; err != nil {
		response.NotFound(c, "主题不存在")
		return
	}

	// 序列化配置
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		response.InternalError(c, "配置格式错误")
		return
	}

	// 更新配置
	if err := db.Model(&theme).Update("config", string(configJSON)).Error; err != nil {
		response.InternalError(c, "更新配置失败")
		return
	}

	response.Success(c, nil)
}

// SetDefaultTheme 设置默认主题
// @Summary 设置默认主题
// @Description 将指定主题设为默认主题
// @Tags 管理员-主题
// @Produce json
// @Security Bearer
// @Param id path int true "主题ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/themes/{id}/default [put]
func (h *ThemeHandler) SetDefaultTheme(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的主题ID")
		return
	}

	db := database.Get()

	// 开始事务
	tx := db.Begin()

	// 取消所有默认主题
	if err := tx.Model(&model.Theme{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		tx.Rollback()
		response.InternalError(c, "设置默认主题失败")
		return
	}

	// 设置新的默认主题
	if err := tx.Model(&model.Theme{}).Where("id = ?", id).Update("is_default", true).Error; err != nil {
		tx.Rollback()
		response.InternalError(c, "设置默认主题失败")
		return
	}

	tx.Commit()

	response.Success(c, nil)
}

// DeleteTheme 删除主题
// @Summary 删除主题
// @Description 删除指定主题
// @Tags 管理员-主题
// @Produce json
// @Security Bearer
// @Param id path int true "主题ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/themes/{id} [delete]
func (h *ThemeHandler) DeleteTheme(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的主题ID")
		return
	}

	db := database.Get()
	var theme model.Theme
	if err := db.First(&theme, id).Error; err != nil {
		response.NotFound(c, "主题不存在")
		return
	}

	// 不能删除默认主题
	if theme.IsDefault {
		response.BadRequest(c, "不能删除默认主题")
		return
	}

	// 删除主题文件
	themePath := filepath.Join(h.themeDir, theme.Name)
	os.RemoveAll(themePath)

	// 删除数据库记录
	if err := db.Delete(&theme).Error; err != nil {
		response.InternalError(c, "删除主题失败")
		return
	}

	response.Success(c, nil)
}

// GetPublicTheme 获取公开主题信息
// @Summary 获取公开主题
// @Description 获取当前使用的主题信息（用户端）
// @Tags 主题
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/themes/current [get]
func (h *ThemeHandler) GetPublicTheme(c *gin.Context) {
	db := database.Get()

	var theme model.Theme
	if err := db.Where("is_default = ? AND status = ?", true, 1).First(&theme).Error; err != nil {
		// 如果没有默认主题，获取第一个启用的主题
		if err := db.Where("status = ?", 1).First(&theme).Error; err != nil {
			response.Success(c, nil)
			return
		}
	}

	// 解析配置
	var config map[string]interface{}
	if theme.Config != "" {
		json.Unmarshal([]byte(theme.Config), &config)
	}

	response.Success(c, gin.H{
		"name":   theme.Name,
		"title":  theme.Title,
		"config": config,
	})
}
