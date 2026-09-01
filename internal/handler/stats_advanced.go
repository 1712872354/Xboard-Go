package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"
)

// StatsAdvancedHandler 统计报表增强处理器
type StatsAdvancedHandler struct{}

// NewStatsAdvancedHandler 创建统计报表增强处理器
func NewStatsAdvancedHandler() *StatsAdvancedHandler {
	return &StatsAdvancedHandler{}
}

// GetServerRanking 获取服务器排行
// @Summary 服务器排行
// @Description 获取服务器流量使用排行
// @Tags 管理员-统计
// @Produce json
// @Security Bearer
// @Param days query int false "天数" default(7)
// @Param limit query int false "数量" default(10)
// @Success 200 {object} response.Response
// @Router /api/v1/admin/stats/server-ranking [get]
func (h *StatsAdvancedHandler) GetServerRanking(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if days < 1 || days > 30 {
		days = 7
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	db := database.Get()

	// 获取服务器流量统计
	type ServerTraffic struct {
		NodeID     uint   `json:"node_id"`
		NodeName   string `json:"node_name"`
		TotalUp    int64  `json:"total_up"`
		TotalDown  int64  `json:"total_down"`
		Total      int64  `json:"total"`
	}

	var stats []ServerTraffic
	startDate := time.Now().AddDate(0, 0, -days)

	db.Model(&model.ServerStat{}).
		Select("node_id, node_name, SUM(upload) as total_up, SUM(download) as total_down, SUM(upload + download) as total").
		Where("record_at >= ?", startDate).
		Group("node_id, node_name").
		Order("total DESC").
		Limit(limit).
		Find(&stats)

	if stats == nil {
		stats = []ServerTraffic{}
	}

	response.Success(c, stats)
}

// GetServerYesterdayRanking 获取服务器昨日排行
// @Summary 服务器昨日排行
// @Description 获取服务器昨日流量使用排行
// @Tags 管理员-统计
// @Produce json
// @Security Bearer
// @Param limit query int false "数量" default(10)
// @Success 200 {object} response.Response
// @Router /api/v1/admin/stats/server-yesterday-ranking [get]
func (h *StatsAdvancedHandler) GetServerYesterdayRanking(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	db := database.Get()

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	type ServerTraffic struct {
		NodeID     uint   `json:"node_id"`
		NodeName   string `json:"node_name"`
		TotalUp    int64  `json:"total_up"`
		TotalDown  int64  `json:"total_down"`
		Total      int64  `json:"total"`
	}

	var stats []ServerTraffic
	db.Model(&model.ServerStat{}).
		Select("node_id, node_name, SUM(upload) as total_up, SUM(download) as total_down, SUM(upload + download) as total").
		Where("DATE(record_at) = ?", yesterday).
		Group("node_id, node_name").
		Order("total DESC").
		Limit(limit).
		Find(&stats)

	if stats == nil {
		stats = []ServerTraffic{}
	}

	response.Success(c, stats)
}

// GetOrderStats 获取订单统计
// @Summary 订单统计
// @Description 获取订单统计数据
// @Tags 管理员-统计
// @Produce json
// @Security Bearer
// @Param days query int false "天数" default(30)
// @Success 200 {object} response.Response
// @Router /api/v1/admin/stats/orders [get]
func (h *StatsAdvancedHandler) GetOrderStats(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	db := database.Get()
	startDate := time.Now().AddDate(0, 0, -days)

	// 总订单数
	var totalOrders int64
	db.Model(&model.Order{}).Where("created_at >= ?", startDate).Count(&totalOrders)

	// 已支付订单数
	var paidOrders int64
	db.Model(&model.Order{}).Where("created_at >= ? AND status = ?", startDate, model.OrderStatusPaid).Count(&paidOrders)

	// 待支付订单数
	var pendingOrders int64
	db.Model(&model.Order{}).Where("created_at >= ? AND status = ?", startDate, model.OrderStatusPending).Count(&pendingOrders)

	// 已取消订单数
	var cancelledOrders int64
	db.Model(&model.Order{}).Where("created_at >= ? AND status = ?", startDate, model.OrderStatusCancelled).Count(&cancelledOrders)

	// 总收入
	var totalIncome float64
	db.Model(&model.Order{}).
		Where("created_at >= ? AND status = ?", startDate, model.OrderStatusPaid).
		Select("COALESCE(SUM(actual_amount), 0)").
		Scan(&totalIncome)

	// 今日订单数
	var todayOrders int64
	today := time.Now().Format("2006-01-02")
	db.Model(&model.Order{}).Where("DATE(created_at) = ?", today).Count(&todayOrders)

	// 今日收入
	var todayIncome float64
	db.Model(&model.Order{}).
		Where("DATE(created_at) = ? AND status = ?", today, model.OrderStatusPaid).
		Select("COALESCE(SUM(actual_amount), 0)").
		Scan(&todayIncome)

	// 每日订单统计
	type DailyOrder struct {
		Date   string  `json:"date"`
		Count  int64   `json:"count"`
		Amount float64 `json:"amount"`
	}

	var dailyOrders []DailyOrder
	rows, err := db.Model(&model.Order{}).
		Select("DATE(created_at) as date, COUNT(*) as count, COALESCE(SUM(actual_amount), 0) as amount").
		Where("created_at >= ?", startDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var daily DailyOrder
			rows.Scan(&daily.Date, &daily.Count, &daily.Amount)
			dailyOrders = append(dailyOrders, daily)
		}
	}

	if dailyOrders == nil {
		dailyOrders = []DailyOrder{}
	}

	response.Success(c, gin.H{
		"total_orders":     totalOrders,
		"paid_orders":      paidOrders,
		"pending_orders":   pendingOrders,
		"cancelled_orders": cancelledOrders,
		"total_income":     totalIncome,
		"today_orders":     todayOrders,
		"today_income":     todayIncome,
		"daily_orders":     dailyOrders,
	})
}

// GetUserStats 获取用户统计
// @Summary 用户统计
// @Description 获取用户统计数据
// @Tags 管理员-统计
// @Produce json
// @Security Bearer
// @Param days query int false "天数" default(30)
// @Success 200 {object} response.Response
// @Router /api/v1/admin/stats/users [get]
func (h *StatsAdvancedHandler) GetUserStats(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	db := database.Get()

	// 总用户数
	var totalUsers int64
	db.Model(&model.User{}).Count(&totalUsers)

	// 活跃用户数（已登录）
	var activeUsers int64
	db.Model(&model.User{}).Where("last_login_at IS NOT NULL").Count(&activeUsers)

	// 今日新增用户
	var todayUsers int64
	today := time.Now().Format("2006-01-02")
	db.Model(&model.User{}).Where("DATE(created_at) = ?", today).Count(&todayUsers)

	// 本周新增用户
	var weekUsers int64
	weekAgo := time.Now().AddDate(0, 0, -7)
	db.Model(&model.User{}).Where("created_at >= ?", weekAgo).Count(&weekUsers)

	// 本月新增用户
	var monthUsers int64
	monthAgo := time.Now().AddDate(0, -1, 0)
	db.Model(&model.User{}).Where("created_at >= ?", monthAgo).Count(&monthUsers)

	// 付费用户数
	var paidUsers int64
	db.Model(&model.User{}).Where("plan_id IS NOT NULL").Count(&paidUsers)

	// 每日新增用户统计
	type DailyUser struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	var dailyUsers []DailyUser
	startDate := time.Now().AddDate(0, 0, -days)
	rows, err := db.Model(&model.User{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", startDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var daily DailyUser
			rows.Scan(&daily.Date, &daily.Count)
			dailyUsers = append(dailyUsers, daily)
		}
	}

	if dailyUsers == nil {
		dailyUsers = []DailyUser{}
	}

	response.Success(c, gin.H{
		"total_users":  totalUsers,
		"active_users": activeUsers,
		"today_users":  todayUsers,
		"week_users":   weekUsers,
		"month_users":  monthUsers,
		"paid_users":   paidUsers,
		"daily_users":  dailyUsers,
	})
}

// GetStatRecords 获取统计记录
// @Summary 统计记录
// @Description 获取详细的统计记录
// @Tags 管理员-统计
// @Produce json
// @Security Bearer
// @Param start_date query string true "开始日期"
// @Param end_date query string true "结束日期"
// @Param type query string false "类型 (user/server/order)"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/stats/records [get]
func (h *StatsAdvancedHandler) GetStatRecords(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	statType := c.DefaultQuery("type", "user")

	if startDate == "" || endDate == "" {
		response.BadRequest(c, "开始日期和结束日期不能为空")
		return
	}

	db := database.Get()

	switch statType {
	case "user":
		type UserStat struct {
			Date        string `json:"date"`
			NewUsers    int64  `json:"new_users"`
			ActiveUsers int64  `json:"active_users"`
		}

		var stats []UserStat
		rows, err := db.Model(&model.User{}).
			Select("DATE(created_at) as date, COUNT(*) as new_users, 0 as active_users").
			Where("DATE(created_at) BETWEEN ? AND ?", startDate, endDate).
			Group("DATE(created_at)").
			Order("date ASC").
			Rows()
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var stat UserStat
				rows.Scan(&stat.Date, &stat.NewUsers, &stat.ActiveUsers)
				stats = append(stats, stat)
			}
		}

		if stats == nil {
			stats = []UserStat{}
		}

		response.Success(c, stats)

	case "server":
		type ServerStat struct {
			Date      string `json:"date"`
			NodeID    uint   `json:"node_id"`
			Upload    int64  `json:"upload"`
			Download  int64  `json:"download"`
		}

		var stats []ServerStat
		rows, err := db.Model(&model.ServerStat{}).
			Select("DATE(record_at) as date, node_id, SUM(upload) as upload, SUM(download) as download").
			Where("DATE(record_at) BETWEEN ? AND ?", startDate, endDate).
			Group("DATE(record_at), node_id").
			Order("date ASC").
			Rows()
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var stat ServerStat
				rows.Scan(&stat.Date, &stat.NodeID, &stat.Upload, &stat.Download)
				stats = append(stats, stat)
			}
		}

		if stats == nil {
			stats = []ServerStat{}
		}

		response.Success(c, stats)

	case "order":
		type OrderStat struct {
			Date      string  `json:"date"`
			Count     int64   `json:"count"`
			Amount    float64 `json:"amount"`
		}

		var stats []OrderStat
		rows, err := db.Model(&model.Order{}).
			Select("DATE(created_at) as date, COUNT(*) as count, COALESCE(SUM(actual_amount), 0) as amount").
			Where("DATE(created_at) BETWEEN ? AND ?", startDate, endDate).
			Group("DATE(created_at)").
			Order("date ASC").
			Rows()
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var stat OrderStat
				rows.Scan(&stat.Date, &stat.Count, &stat.Amount)
				stats = append(stats, stat)
			}
		}

		if stats == nil {
			stats = []OrderStat{}
		}

		response.Success(c, stats)

	default:
		response.BadRequest(c, "无效的统计类型")
	}
}
