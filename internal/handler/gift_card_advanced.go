package handler

import (
	"encoding/csv"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"
)

// GiftCardAdvancedHandler 礼品卡高级处理器
type GiftCardAdvancedHandler struct{}

// NewGiftCardAdvancedHandler 创建礼品卡高级处理器
func NewGiftCardAdvancedHandler() *GiftCardAdvancedHandler {
	return &GiftCardAdvancedHandler{}
}

// ToggleCodeRequest 启用/禁用卡密请求
type ToggleCodeRequest struct {
	Enabled bool `json:"enabled"`
}

// ToggleCode 启用/禁用卡密
// @Summary 启用/禁用卡密
// @Description 切换卡密的启用状态
// @Tags 管理员-礼品卡
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "卡密ID"
// @Param request body ToggleCodeRequest true "启用状态"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/gift-card-codes/{id}/toggle [put]
func (h *GiftCardAdvancedHandler) ToggleCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的卡密ID")
		return
	}

	var req ToggleCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	db := database.Get()
	var code model.GiftCardCode
	if err := db.First(&code, id).Error; err != nil {
		response.NotFound(c, "卡密不存在")
		return
	}

	// 更新状态
	status := 0
	if req.Enabled {
		status = 1
	}
	if err := db.Model(&code).Update("status", status).Error; err != nil {
		response.InternalError(c, "更新状态失败")
		return
	}

	response.Success(c, nil)
}

// ExportCodes 导出卡密
// @Summary 导出卡密
// @Description 导出卡密为CSV文件
// @Tags 管理员-礼品卡
// @Produce csv
// @Security Bearer
// @Param template_id query int false "模板ID"
// @Param status query int false "状态 (-1=全部, 0=未使用, 1=已使用)"
// @Success 200 {string} string "CSV文件"
// @Router /api/v1/admin/gift-card-codes/export [get]
func (h *GiftCardAdvancedHandler) ExportCodes(c *gin.Context) {
	templateID, _ := strconv.ParseUint(c.Query("template_id"), 10, 32)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	db := database.Get()
	query := db.Model(&model.GiftCardCode{})

	if templateID > 0 {
		query = query.Where("template_id = ?", templateID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	var codes []model.GiftCardCode
	query.Find(&codes)

	// 创建CSV
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=gift_card_codes.csv")

	writer := csv.NewWriter(c.Writer)
	writer.Write([]string{"ID", "卡密", "模板ID", "状态", "使用时间", "创建时间"})

	for _, code := range codes {
		statusStr := "未使用"
		if code.Status == 1 {
			statusStr = "已使用"
		}
		usedAt := ""
		if code.UsedAt != nil {
			usedAt = code.UsedAt.Format("2006-01-02 15:04:05")
		}
		writer.Write([]string{
			strconv.FormatUint(uint64(code.ID), 10),
			code.Code,
			strconv.FormatUint(uint64(code.TemplateID), 10),
			statusStr,
			usedAt,
			code.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	writer.Flush()
}

// UpdateCodeRequest 更新卡密请求
type UpdateCodeRequest struct {
	Code       string `json:"code"`
	TemplateID uint   `json:"template_id"`
}

// UpdateCode 更新卡密
// @Summary 更新卡密
// @Description 更新卡密信息
// @Tags 管理员-礼品卡
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "卡密ID"
// @Param request body UpdateCodeRequest true "卡密信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/gift-card-codes/{id} [put]
func (h *GiftCardAdvancedHandler) UpdateCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的卡密ID")
		return
	}

	var req UpdateCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	db := database.Get()
	var code model.GiftCardCode
	if err := db.First(&code, id).Error; err != nil {
		response.NotFound(c, "卡密不存在")
		return
	}

	// 检查是否已使用
	if code.Status == 1 {
		response.BadRequest(c, "已使用的卡密不能修改")
		return
	}

	updates := map[string]interface{}{}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.TemplateID > 0 {
		updates["template_id"] = req.TemplateID
	}

	if len(updates) > 0 {
		if err := db.Model(&code).Updates(updates).Error; err != nil {
			response.InternalError(c, "更新失败")
			return
		}
	}

	response.Success(c, nil)
}

// GetCodeUsages 获取卡密使用记录
// @Summary 卡密使用记录
// @Description 获取卡密的使用记录
// @Tags 管理员-礼品卡
// @Produce json
// @Security Bearer
// @Param id path int true "卡密ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/gift-card-codes/{id}/usages [get]
func (h *GiftCardAdvancedHandler) GetCodeUsages(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的卡密ID")
		return
	}

	db := database.Get()

	// 查询卡密
	var code model.GiftCardCode
	if err := db.First(&code, id).Error; err != nil {
		response.NotFound(c, "卡密不存在")
		return
	}

	// 查询使用记录
	var usages []model.GiftCardUsage
	db.Where("code_id = ?", id).
		Order("created_at DESC").
		Find(&usages)

	if usages == nil {
		usages = []model.GiftCardUsage{}
	}

	// 获取用户信息
	userIDs := make([]uint, 0)
	for _, usage := range usages {
		userIDs = append(userIDs, usage.UserID)
	}

	var users []model.User
	db.Where("id IN ?", userIDs).Find(&users)
	userMap := make(map[uint]model.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// 构建响应
	var result []gin.H
	for _, usage := range usages {
		user := userMap[usage.UserID]
		result = append(result, gin.H{
			"id":         usage.ID,
			"user_id":    usage.UserID,
			"user_email": user.Email,
			"amount":     usage.Amount,
			"traffic":    usage.Traffic,
			"duration":   usage.Duration,
			"created_at": usage.CreatedAt,
		})
	}

	if result == nil {
		result = []gin.H{}
	}

	response.Success(c, result)
}

// GetStatistics 获取礼品卡统计
// @Summary 礼品卡统计
// @Description 获取礼品卡统计数据
// @Tags 管理员-礼品卡
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/gift-card-codes/statistics [get]
func (h *GiftCardAdvancedHandler) GetStatistics(c *gin.Context) {
	db := database.Get()

	// 总卡密数
	var totalCodes int64
	db.Model(&model.GiftCardCode{}).Count(&totalCodes)

	// 已使用卡密数
	var usedCodes int64
	db.Model(&model.GiftCardCode{}).Where("status = ?", 1).Count(&usedCodes)

	// 未使用卡密数
	unusedCodes := totalCodes - usedCodes

	// 今日使用数
	var todayUsed int64
	today := time.Now().Format("2006-01-02")
	db.Model(&model.GiftCardUsage{}).Where("DATE(created_at) = ?", today).Count(&todayUsed)

	// 按模板统计
	type TemplateStat struct {
		TemplateID   uint   `json:"template_id"`
		TemplateName string `json:"template_name"`
		Total        int64  `json:"total"`
		Used         int64  `json:"used"`
	}

	var templateStats []TemplateStat
	rows, err := db.Model(&model.GiftCardCode{}).
		Select("template_id, COUNT(*) as total, SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as used").
		Group("template_id").
		Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat TemplateStat
			rows.Scan(&stat.TemplateID, &stat.Total, &stat.Used)
			
			// 获取模板名称
			var template model.GiftCardTemplate
			if err := db.First(&template, stat.TemplateID).Error; err == nil {
				stat.TemplateName = template.Name
			}
			
			templateStats = append(templateStats, stat)
		}
	}

	if templateStats == nil {
		templateStats = []TemplateStat{}
	}

	response.Success(c, gin.H{
		"total_codes":    totalCodes,
		"used_codes":     usedCodes,
		"unused_codes":   unusedCodes,
		"today_used":     todayUsed,
		"template_stats": templateStats,
	})
}

// GetTypes 获取礼品卡类型
// @Summary 礼品卡类型
// @Description 获取礼品卡类型列表
// @Tags 管理员-礼品卡
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/gift-card-codes/types [get]
func (h *GiftCardAdvancedHandler) GetTypes(c *gin.Context) {
	types := []gin.H{
		{"id": 1, "name": "余额充值", "description": "充值账户余额"},
		{"id": 2, "name": "流量充值", "description": "充值账户流量"},
		{"id": 3, "name": "时长充值", "description": "延长套餐时长"},
		{"id": 4, "name": "套餐兑换", "description": "兑换指定套餐"},
	}

	response.Success(c, types)
}

// GetUserGiftCardHistory 用户礼品卡历史
// @Summary 用户礼品卡历史
// @Description 获取用户的礼品卡使用历史
// @Tags 用户-礼品卡
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response
// @Router /api/v1/gift-cards/history [get]
func (h *GiftCardAdvancedHandler) GetUserGiftCardHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	db := database.Get()

	var total int64
	db.Model(&model.GiftCardUsage{}).Where("user_id = ?", userID).Count(&total)

	var usages []model.GiftCardUsage
	offset := (page - 1) * pageSize
	db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&usages)

	if usages == nil {
		usages = []model.GiftCardUsage{}
	}

	// 获取卡密和模板信息
	codeIDs := make([]uint, 0)
	for _, usage := range usages {
		codeIDs = append(codeIDs, usage.CodeID)
	}

	var codes []model.GiftCardCode
	db.Where("id IN ?", codeIDs).Find(&codes)
	codeMap := make(map[uint]model.GiftCardCode)
	for _, code := range codes {
		codeMap[code.ID] = code
	}

	templateIDs := make([]uint, 0)
	for _, code := range codes {
		templateIDs = append(templateIDs, code.TemplateID)
	}

	var templates []model.GiftCardTemplate
	db.Where("id IN ?", templateIDs).Find(&templates)
	templateMap := make(map[uint]model.GiftCardTemplate)
	for _, template := range templates {
		templateMap[template.ID] = template
	}

	// 构建响应
	var result []gin.H
	for _, usage := range usages {
		code := codeMap[usage.CodeID]
		template := templateMap[code.TemplateID]
		result = append(result, gin.H{
			"id":             usage.ID,
			"code":           code.Code,
			"template_name":  template.Name,
			"value":          template.Value,
			"type":           template.Type,
			"created_at":     usage.CreatedAt,
		})
	}

	if result == nil {
		result = []gin.H{}
	}

	response.Success(c, gin.H{
		"list":      result,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetUserGiftCardDetail 用户礼品卡详情
// @Summary 用户礼品卡详情
// @Description 获取礼品卡详情
// @Tags 用户-礼品卡
// @Produce json
// @Security Bearer
// @Param id path int true "使用记录ID"
// @Success 200 {object} response.Response
// @Router /api/v1/gift-cards/{id} [get]
func (h *GiftCardAdvancedHandler) GetUserGiftCardDetail(c *gin.Context) {
	userID, _ := c.Get("user_id")
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的记录ID")
		return
	}

	db := database.Get()
	var usage model.GiftCardUsage
	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&usage).Error; err != nil {
		response.NotFound(c, "记录不存在")
		return
	}

	// 获取卡密信息
	var code model.GiftCardCode
	db.First(&code, usage.CodeID)

	// 获取模板信息
	var template model.GiftCardTemplate
	db.First(&template, code.TemplateID)

	response.Success(c, gin.H{
		"id":             usage.ID,
		"code":           code.Code,
		"template_name":  template.Name,
		"description":    template.Description,
		"value":          template.Value,
		"traffic":        template.Traffic,
		"duration":       template.Duration,
		"type":           template.Type,
		"created_at":     usage.CreatedAt,
	})
}

// GetUserGiftCardTypes 用户礼品卡类型
// @Summary 用户礼品卡类型
// @Description 获取礼品卡类型列表（用户端）
// @Tags 用户-礼品卡
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/gift-cards/types [get]
func (h *GiftCardAdvancedHandler) GetUserGiftCardTypes(c *gin.Context) {
	types := []gin.H{
		{"id": 1, "name": "余额充值", "description": "充值账户余额"},
		{"id": 2, "name": "流量充值", "description": "充值账户流量"},
		{"id": 3, "name": "时长充值", "description": "延长套餐时长"},
		{"id": 4, "name": "套餐兑换", "description": "兑换指定套餐"},
	}

	response.Success(c, types)
}
