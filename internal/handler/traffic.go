package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/middleware"
	"xboard-go/internal/service"
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
