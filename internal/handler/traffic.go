package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/middleware"
	"xboard-go/internal/model"
	"xboard-go/internal/service"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"
)

// TrafficHandler 流量处理器
type TrafficHandler struct {
	trafficService service.TrafficService
}

// NewTrafficHandler 创建流量处理器
func NewTrafficHandler(trafficService service.TrafficService) *TrafficHandler {
	return &TrafficHandler{
		trafficService: trafficService,
	}
}

// GetStats 获取流量统计
// @Summary 获取流量统计
// @Description 获取当前用户的流量统计（今日/本周/本月/总量）
// @Tags 流量
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=service.TrafficStats}
// @Router /api/v1/user/traffic/stats [get]
func (h *TrafficHandler) GetStats(c *gin.Context) {
	userID := middleware.GetUserID(c)

	stats, err := h.trafficService.GetUserTrafficStats(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

// GetHistory 获取流量历史记录
// @Summary 获取流量历史
// @Description 获取当前用户的流量使用历史记录
// @Tags 流量
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param start query string false "开始时间 (RFC3339格式)"
// @Param end query string false "结束时间 (RFC3339格式)"
// @Success 200 {object} response.Response
// @Router /api/v1/user/traffic/history [get]
func (h *TrafficHandler) GetHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var start, end time.Time
	if startStr := c.Query("start"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			start = t
		}
	}
	if endStr := c.Query("end"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			end = t
		}
	}

	logs, total, err := h.trafficService.GetUserTrafficHistory(
		c.Request.Context(), userID, page, pageSize, start, end,
	)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetDailyTraffic 获取按日聚合的流量数据
// @Summary 获取每日流量统计
// @Description 获取当前用户按日聚合的流量数据（用于图表展示）
// @Tags 流量
// @Produce json
// @Security Bearer
// @Param days query int false "天数" default(7)
// @Param start_date query string false "开始日期 (YYYY-MM-DD)"
// @Param end_date query string false "结束日期 (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=[]service.DailyTraffic}
// @Router /api/v1/user/traffic/daily [get]
func (h *TrafficHandler) GetDailyTraffic(c *gin.Context) {
	userID := middleware.GetUserID(c)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if endDateStr := c.Query("end_date"); endDateStr != "" {
			startDate, err1 := time.Parse("2006-01-02", startDateStr)
			endDate, err2 := time.Parse("2006-01-02", endDateStr)
			if err1 == nil && err2 == nil {
				daily, err := h.trafficService.GetDailyTrafficByDateRange(
					c.Request.Context(), userID, startDate, endDate,
				)
				if err != nil {
					response.InternalError(c, err.Error())
					return
				}
				response.Success(c, daily)
				return
			}
		}
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	daily, err := h.trafficService.GetDailyTraffic(c.Request.Context(), userID, days)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, daily)
}

// SyncTraffic 手动触发流量同步（管理员）
// @Summary 同步流量数据
// @Description 手动触发将Redis中的流量数据同步到数据库（管理员权限）
// @Tags 管理员-流量
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/traffic/sync [post]
func (h *TrafficHandler) SyncTraffic(c *gin.Context) {
	synced, err := h.trafficService.SyncAllToDB(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"synced_count": synced,
	})
}

// AdminGetUserTraffic 获取指定用户的流量统计（管理员）
// @Summary 获取用户流量统计
// @Description 获取指定用户的流量统计（管理员权限）
// @Tags 管理员-流量
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=service.TrafficStats}
// @Router /api/v1/admin/users/{id}/traffic [get]
func (h *TrafficHandler) AdminGetUserTraffic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	stats, err := h.trafficService.GetUserTrafficStats(c.Request.Context(), uint(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

// TrafficResetLogWithUser 带用户邮箱的流量重置日志
type TrafficResetLogWithUser struct {
	model.TrafficResetLog
	UserEmail string `json:"user_email"`
}

// ListTrafficLogs 流量重置日志列表（管理员）
func (h *TrafficHandler) ListTrafficLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	db := database.Get()
	query := db.Model(&model.TrafficResetLog{})

	// 用户邮箱筛选
	if userEmail := c.Query("user_email"); userEmail != "" {
		var userIDs []uint
		db.Model(&model.User{}).Where("email LIKE ?", "%"+userEmail+"%").Pluck("id", &userIDs)
		if len(userIDs) > 0 {
			query = query.Where("user_id IN ?", userIDs)
		} else {
			response.Success(c, gin.H{"list": []TrafficResetLogWithUser{}, "total": 0})
			return
		}
	}

	// 重置类型筛选
	if resetType := c.Query("reset_type"); resetType != "" {
		query = query.Where("reset_type = ?", resetType)
	}

	// 触发来源筛选
	if triggerSource := c.Query("trigger_source"); triggerSource != "" {
		query = query.Where("trigger_source = ?", triggerSource)
	}

	// 时间范围筛选
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("reset_time >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("reset_time <= ?", endDate+" 23:59:59")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.InternalError(c, "查询日志失败")
		return
	}

	var logs []model.TrafficResetLog
	offset := (page - 1) * perPage
	if err := query.Order("id DESC").Offset(offset).Limit(perPage).Find(&logs).Error; err != nil {
		response.InternalError(c, "查询日志失败")
		return
	}

	// 批量查询用户邮箱
	userIDSet := make(map[uint]bool)
	for _, log := range logs {
		userIDSet[log.UserID] = true
	}
	userIDs := make([]uint, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	userEmailMap := make(map[uint]string)
	if len(userIDs) > 0 {
		var users []model.User
		db.Where("id IN ?", userIDs).Find(&users)
		for _, u := range users {
			userEmailMap[u.ID] = u.Email
		}
	}

	// 组装结果
	result := make([]TrafficResetLogWithUser, len(logs))
	for i, log := range logs {
		result[i] = TrafficResetLogWithUser{
			TrafficResetLog: log,
			UserEmail:       userEmailMap[log.UserID],
		}
	}

	response.Success(c, gin.H{
		"list":  result,
		"total": total,
	})
}
