package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"
)

// TrafficResetHandler 流量重置处理器
type TrafficResetHandler struct{}

// NewTrafficResetHandler 创建流量重置处理器
func NewTrafficResetHandler() *TrafficResetHandler {
	return &TrafficResetHandler{}
}

// GetTrafficResetLogs 获取流量重置日志
// @Summary 流量重置日志
// @Description 获取流量重置日志列表
// @Tags 管理员-流量
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param user_id query int false "用户ID"
// @Param reset_type query string false "重置类型"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/traffic-reset/logs [get]
func (h *TrafficResetHandler) GetTrafficResetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 32)
	resetType := c.Query("reset_type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	db := database.Get()
	query := db.Model(&model.TrafficResetLog{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if resetType != "" {
		query = query.Where("reset_type = ?", resetType)
	}

	var total int64
	query.Count(&total)

	var logs []model.TrafficResetLog
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs)

	if logs == nil {
		logs = []model.TrafficResetLog{}
	}

	// 获取用户信息
	userIDs := make([]uint, 0)
	for _, log := range logs {
		userIDs = append(userIDs, log.UserID)
	}

	var users []model.User
	db.Where("id IN ?", userIDs).Find(&users)
	userMap := make(map[uint]model.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// 构建响应
	var result []gin.H
	for _, log := range logs {
		user := userMap[log.UserID]
		result = append(result, gin.H{
			"id":          log.ID,
			"user_id":     log.UserID,
			"user_email":  user.Email,
			"reset_type":  log.ResetType,
			"old_traffic": log.OldTotal,
			"new_traffic": log.NewTotal,
			"created_at":  log.CreatedAt,
		})
	}

	if result == nil {
		result = []gin.H{}
	}

	response.Success(c, gin.H{
		"list":  result,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// GetTrafficResetStats 获取流量重置统计
// @Summary 流量重置统计
// @Description 获取流量重置统计数据
// @Tags 管理员-流量
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/traffic-reset/stats [get]
func (h *TrafficResetHandler) GetTrafficResetStats(c *gin.Context) {
	db := database.Get()

	// 今日重置次数
	var todayCount int64
	today := time.Now().Format("2006-01-02")
	db.Model(&model.TrafficResetLog{}).Where("DATE(created_at) = ?", today).Count(&todayCount)

	// 本周重置次数
	var weekCount int64
	weekAgo := time.Now().AddDate(0, 0, -7)
	db.Model(&model.TrafficResetLog{}).Where("created_at >= ?", weekAgo).Count(&weekCount)

	// 本月重置次数
	var monthCount int64
	monthAgo := time.Now().AddDate(0, -1, 0)
	db.Model(&model.TrafficResetLog{}).Where("created_at >= ?", monthAgo).Count(&monthCount)

	// 总重置次数
	var totalCount int64
	db.Model(&model.TrafficResetLog{}).Count(&totalCount)

	// 按类型统计
	typeStats := []gin.H{}
	rows, err := db.Model(&model.TrafficResetLog{}).
		Select("reset_type, COUNT(*) as count").
		Group("reset_type").
		Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var resetType string
			var count int64
			rows.Scan(&resetType, &count)
			typeStats = append(typeStats, gin.H{
				"type":  resetType,
				"count": count,
			})
		}
	}

	if typeStats == nil {
		typeStats = []gin.H{}
	}

	response.Success(c, gin.H{
		"today":      todayCount,
		"week":       weekCount,
		"month":      monthCount,
		"total":      totalCount,
		"type_stats": typeStats,
	})
}

// GetUserTrafficResetHistory 获取用户流量重置历史
// @Summary 用户流量重置历史
// @Description 获取指定用户的流量重置历史
// @Tags 管理员-流量
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response
// @Router /api/v1/admin/traffic-reset/user/{id}/history [get]
func (h *TrafficResetHandler) GetUserTrafficResetHistory(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

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
	db.Model(&model.TrafficResetLog{}).Where("user_id = ?", userID).Count(&total)

	var logs []model.TrafficResetLog
	offset := (page - 1) * pageSize
	db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs)

	if logs == nil {
		logs = []model.TrafficResetLog{}
	}

	response.Success(c, gin.H{
		"list":  logs,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// ResetUserTrafficRequest 重置用户流量请求
type ResetUserTrafficRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

// ResetUserTraffic 手动重置用户流量
// @Summary 重置用户流量
// @Description 手动重置指定用户的流量
// @Tags 管理员-流量
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body ResetUserTrafficRequest true "用户信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/traffic-reset/reset-user [post]
func (h *TrafficResetHandler) ResetUserTraffic(c *gin.Context) {
	var req ResetUserTrafficRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	db := database.Get()

	// 获取用户
	var user model.User
	if err := db.First(&user, req.UserID).Error; err != nil {
		response.BadRequest(c, "用户不存在")
		return
	}

	// 记录旧流量
	oldTraffic := user.UsedTraffic

	// 开始事务
	tx := db.Begin()

	// 重置流量
	if err := tx.Model(&user).Update("used_traffic", 0).Error; err != nil {
		tx.Rollback()
		response.InternalError(c, "重置流量失败")
		return
	}

	// 记录重置日志
	log := model.TrafficResetLog{
		UserID:     req.UserID,
		ResetType:  "manual",
		OldTotal: oldTraffic,
		NewTotal: 0,
		ResetTime:  time.Now(),
	}
	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		response.InternalError(c, "记录重置日志失败")
		return
	}

	tx.Commit()

	response.Success(c, gin.H{
		"message":     "流量重置成功",
		"old_traffic": oldTraffic,
		"new_traffic": 0,
	})
}
