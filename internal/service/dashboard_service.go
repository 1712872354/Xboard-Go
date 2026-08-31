package service

import (
	"sync"
	"time"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
	"xboard-go/pkg/database"
)

// DashboardService 看板服务接口
type DashboardService interface {
	// GetOverview 获取概览数据
	GetOverview() (*DashboardOverview, error)
	// GetRecentOrders 获取最近订单
	GetRecentOrders(limit int) ([]model.Order, error)
	// GetRecentUsers 获取最近注册用户
	GetRecentUsers(limit int) ([]model.User, error)
	// GetIncomeStats 获取收入统计（最近N天）
	GetIncomeStats(days int) ([]DailyIncome, error)
	// GetUserGrowthStats 获取用户增长统计（最近N天）
	GetUserGrowthStats(days int) ([]DailyUserGrowth, error)
	// GetNodeStats 获取节点统计
	GetNodeStats() (*NodeStats, error)
	// GetNodeTrafficRanking 获取节点流量排行
	GetNodeTrafficRanking(limit int) ([]NodeTrafficRank, error)
	// GetUserTrafficRanking 获取用户流量消耗排行
	GetUserTrafficRanking(limit int) ([]UserTrafficRank, error)
	// GetInviteRanking 获取邀请排行
	GetInviteRanking(limit int) ([]InviteRank, error)
	// GetCommissionStats 获取佣金统计
	GetCommissionStats() (*CommissionStats, error)
	// GetComprehensiveStats 获取综合统计
	GetComprehensiveStats() (*ComprehensiveStats, error)
}

type dashboardService struct {
	userRepo   repository.UserRepository
	orderRepo  repository.OrderRepository
	planRepo   repository.PlanRepository
	nodeRepo   repository.NodeRepository
	ticketRepo repository.TicketRepository
	redeemRepo repository.RedeemCodeRepository

	// 缓存
	overviewCache     *DashboardOverview
	overviewCacheTime time.Time
	overviewCacheMu   sync.RWMutex
}

// DashboardOverview 看板概览
type DashboardOverview struct {
	TotalUsers    int64   `json:"total_users"`
	ActiveUsers   int64   `json:"active_users"`
	TotalOrders   int64   `json:"total_orders"`
	PaidOrders    int64   `json:"paid_orders"`
	TotalIncome   float64 `json:"total_income"`
	TodayIncome   float64 `json:"today_income"`
	TotalNodes    int64   `json:"total_nodes"`
	OnlineNodes   int64   `json:"online_nodes"`
	OpenTickets   int64   `json:"open_tickets"`
	UnusedRedeems int64   `json:"unused_redeems"`
}

// DailyIncome 每日收入
type DailyIncome struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
	Count  int64   `json:"count"`
}

// DailyUserGrowth 每日用户增长
type DailyUserGrowth struct {
	Date     string `json:"date"`
	NewUsers int64  `json:"new_users"`
}

// NodeStats 节点统计
type NodeStats struct {
	Total       int64 `json:"total"`
	Online      int64 `json:"online"`
	Offline     int64 `json:"offline"`
	Maintenance int64 `json:"maintenance"`
}

// NodeTrafficRank 节点流量排行
type NodeTrafficRank struct {
	NodeID    uint   `json:"node_id"`
	NodeName  string `json:"node_name"`
	Upload    int64  `json:"upload"`
	Download  int64  `json:"download"`
	Total     int64  `json:"total"`
	UserCount int64  `json:"user_count"`
}

// UserTrafficRank 用户流量消耗排行
type UserTrafficRank struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
	Total    int64  `json:"total"`
}

// InviteRank 邀请排行
type InviteRank struct {
	UserID      uint    `json:"user_id"`
	Email       string  `json:"email"`
	InviteCount int64   `json:"invite_count"`
	Commission  float64 `json:"commission"`
}

// CommissionStats 佣金统计
type CommissionStats struct {
	TotalCommission   float64 `json:"total_commission"`
	SettledCommission float64 `json:"settled_commission"`
	PendingCommission float64 `json:"pending_commission"`
	TotalInvites      int64   `json:"total_invites"`
}

// ComprehensiveStats 综合统计
type ComprehensiveStats struct {
	// 收入相关
	TodayIncome         float64 `json:"today_income"`
	DayIncomeGrowth     float64 `json:"day_income_growth"`
	CurrentMonthIncome  float64 `json:"current_month_income"`
	LastMonthIncome     float64 `json:"last_month_income"`
	MonthIncomeGrowth   float64 `json:"month_income_growth"`

	// 佣金相关
	CurrentMonthCommissionPayout float64 `json:"current_month_commission_payout"`
	LastMonthCommissionPayout    float64 `json:"last_month_commission_payout"`
	CommissionGrowth             float64 `json:"commission_growth"`
	CommissionPendingTotal       int64   `json:"commission_pending_total"`

	// 用户相关
	CurrentMonthNewUsers int64 `json:"current_month_new_users"`
	TotalUsers           int64 `json:"total_users"`
	ActiveUsers          int64 `json:"active_users"`
	UserGrowth           float64 `json:"user_growth"`
	OnlineUsers          int64 `json:"online_users"`
	OnlineDevices        int64 `json:"online_devices"`

	// 工单相关
	TicketPendingTotal int64 `json:"ticket_pending_total"`

	// 节点相关
	OnlineNodes int64 `json:"online_nodes"`

	// 流量统计
	TodayTraffic TrafficSummary `json:"today_traffic"`
	MonthTraffic TrafficSummary `json:"month_traffic"`
	TotalTraffic TrafficSummary `json:"total_traffic"`
}

// TrafficSummary 流量汇总
type TrafficSummary struct {
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
	Total    int64 `json:"total"`
}

// NewDashboardService 创建看板服务
func NewDashboardService(
	userRepo repository.UserRepository,
	orderRepo repository.OrderRepository,
	planRepo repository.PlanRepository,
	nodeRepo repository.NodeRepository,
	ticketRepo repository.TicketRepository,
	redeemRepo repository.RedeemCodeRepository,
) DashboardService {
	return &dashboardService{
		userRepo:   userRepo,
		orderRepo:  orderRepo,
		planRepo:   planRepo,
		nodeRepo:   nodeRepo,
		ticketRepo: ticketRepo,
		redeemRepo: redeemRepo,
	}
}

// GetOverview 获取概览数据
func (s *dashboardService) GetOverview() (*DashboardOverview, error) {
	// 检查缓存（5分钟有效期）
	s.overviewCacheMu.RLock()
	if s.overviewCache != nil && time.Since(s.overviewCacheTime) < 5*time.Minute {
		defer s.overviewCacheMu.RUnlock()
		return s.overviewCache, nil
	}
	s.overviewCacheMu.RUnlock()

	// 缓存过期或不存在，查询数据库
	s.overviewCacheMu.Lock()
	defer s.overviewCacheMu.Unlock()

	// 双重检查
	if s.overviewCache != nil && time.Since(s.overviewCacheTime) < 5*time.Minute {
		return s.overviewCache, nil
	}

	db := database.Get()
	overview := &DashboardOverview{}

	// 用户总数
	var totalUsers int64
	if err := db.Model(&model.User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}
	overview.TotalUsers = totalUsers

	// 活跃用户（有流量记录的用户）
	var activeUsers int64
	if err := db.Model(&model.TrafficLog{}).
		Select("COUNT(DISTINCT user_id)").
		Where("recorded_at >= ?", time.Now().AddDate(0, 0, -7)).
		Scan(&activeUsers).Error; err != nil {
		// 表可能不存在，默认0
		activeUsers = 0
	}
	overview.ActiveUsers = activeUsers

	// 订单总数
	var totalOrders int64
	if err := db.Model(&model.Order{}).Count(&totalOrders).Error; err != nil {
		return nil, err
	}
	overview.TotalOrders = totalOrders

	// 已支付订单数
	var paidOrders int64
	if err := db.Model(&model.Order{}).Where("status = ?", model.OrderStatusPaid).Count(&paidOrders).Error; err != nil {
		return nil, err
	}
	overview.PaidOrders = paidOrders

	// 总收入
	var totalIncome float64
	if err := db.Model(&model.Order{}).
		Where("status = ?", model.OrderStatusPaid).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalIncome).Error; err != nil {
		return nil, err
	}
	overview.TotalIncome = totalIncome

	// 今日收入
	todayStart := time.Now().Truncate(24 * time.Hour)
	var todayIncome float64
	if err := db.Model(&model.Order{}).
		Where("status = ? AND paid_at >= ?", model.OrderStatusPaid, todayStart).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&todayIncome).Error; err != nil {
		return nil, err
	}
	overview.TodayIncome = todayIncome

	// 节点总数
	var totalNodes int64
	if err := db.Model(&model.Node{}).Where("status = ?", 1).Count(&totalNodes).Error; err != nil {
		return nil, err
	}
	overview.TotalNodes = totalNodes

	// 在线节点（最近5分钟有心跳）
	var onlineNodes int64
	if err := db.Model(&model.Node{}).
		Where("status = ? AND last_online >= ?", 1, time.Now().Add(-5*time.Minute)).
		Count(&onlineNodes).Error; err != nil {
		onlineNodes = 0
	}
	overview.OnlineNodes = onlineNodes

	// 待处理工单
	openTickets, err := s.ticketRepo.CountByStatus(model.TicketStatusOpen)
	if err != nil {
		return nil, err
	}
	overview.OpenTickets = openTickets

	// 未使用兑换码
	unusedRedeems, err := s.redeemRepo.CountByStatus(0)
	if err != nil {
		return nil, err
	}
	overview.UnusedRedeems = unusedRedeems

	// 更新缓存
	s.overviewCache = overview
	s.overviewCacheTime = time.Now()

	return overview, nil
}

// GetRecentOrders 获取最近订单
func (s *dashboardService) GetRecentOrders(limit int) ([]model.Order, error) {
	var orders []model.Order
	db := database.Get()

	if err := db.Preload("User").Preload("Plan").
		Order("id DESC").
		Limit(limit).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// GetRecentUsers 获取最近注册用户
func (s *dashboardService) GetRecentUsers(limit int) ([]model.User, error) {
	var users []model.User
	db := database.Get()

	if err := db.Order("id DESC").
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetIncomeStats 获取收入统计（最近N天）
func (s *dashboardService) GetIncomeStats(days int) ([]DailyIncome, error) {
	if days <= 0 || days > 365 {
		days = 7
	}

	db := database.Get()
	startDate := time.Now().AddDate(0, 0, -days+1).Truncate(24 * time.Hour)

	type result struct {
		Date   string
		Amount float64
		Count  int64
	}

	var results []result
	if err := db.Model(&model.Order{}).
		Select("DATE(paid_at) as date, COALESCE(SUM(amount), 0) as amount, COUNT(*) as count").
		Where("status = ? AND paid_at >= ?", model.OrderStatusPaid, startDate).
		Group("DATE(paid_at)").
		Order("date ASC").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	// 补全缺失的日期
	resultMap := make(map[string]*result)
	for i := range results {
		resultMap[results[i].Date] = &results[i]
	}

	var dailyIncomes []DailyIncome
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")

		if r, exists := resultMap[dateStr]; exists {
			dailyIncomes = append(dailyIncomes, DailyIncome{
				Date:   dateStr,
				Amount: r.Amount,
				Count:  r.Count,
			})
		} else {
			dailyIncomes = append(dailyIncomes, DailyIncome{
				Date:   dateStr,
				Amount: 0,
				Count:  0,
			})
		}
	}

	return dailyIncomes, nil
}

// GetUserGrowthStats 获取用户增长统计
func (s *dashboardService) GetUserGrowthStats(days int) ([]DailyUserGrowth, error) {
	if days <= 0 || days > 365 {
		days = 7
	}

	db := database.Get()
	startDate := time.Now().AddDate(0, 0, -days+1).Truncate(24 * time.Hour)

	type result struct {
		Date     string
		NewUsers int64
	}

	var results []result
	if err := db.Model(&model.User{}).
		Select("DATE(created_at) as date, COUNT(*) as new_users").
		Where("created_at >= ?", startDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	// 补全缺失的日期
	resultMap := make(map[string]*result)
	for i := range results {
		resultMap[results[i].Date] = &results[i]
	}

	var growth []DailyUserGrowth
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")

		if r, exists := resultMap[dateStr]; exists {
			growth = append(growth, DailyUserGrowth{
				Date:     dateStr,
				NewUsers: r.NewUsers,
			})
		} else {
			growth = append(growth, DailyUserGrowth{
				Date:     dateStr,
				NewUsers: 0,
			})
		}
	}

	return growth, nil
}

// GetNodeStats 获取节点统计
func (s *dashboardService) GetNodeStats() (*NodeStats, error) {
	db := database.Get()
	stats := &NodeStats{}

	// 总节点数（启用的）
	if err := db.Model(&model.Node{}).Where("status = ?", 1).Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	// 在线节点
	if err := db.Model(&model.Node{}).
		Where("status = ? AND last_online >= ?", 1, time.Now().Add(-5*time.Minute)).
		Count(&stats.Online).Error; err != nil {
		return nil, err
	}

	// 维护中节点（status=2）
	if err := db.Model(&model.Node{}).
		Where("status = ?", 2).
		Count(&stats.Maintenance).Error; err != nil {
		stats.Maintenance = 0
	}

	stats.Offline = stats.Total - stats.Online
	return stats, nil
}

// GetNodeTrafficRanking 获取节点流量排行
func (s *dashboardService) GetNodeTrafficRanking(limit int) ([]NodeTrafficRank, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	db := database.Get()
	var rankings []NodeTrafficRank

	// 查询节点流量排行（基于NodeUser表的流量统计）
	// 由于NodeUser表可能不存在，使用简化查询
	err := db.Model(&model.Node{}).
		Select("nodes.id as node_id, nodes.name as node_name, 0 as upload, 0 as download, 0 as total, 0 as user_count").
		Where("nodes.status = ?", 1).
		Order("nodes.id ASC").
		Limit(limit).
		Scan(&rankings).Error

	if err != nil {
		// 如果查询失败，返回空列表
		return []NodeTrafficRank{}, nil
	}

	return rankings, nil
}

// GetUserTrafficRanking 获取用户流量消耗排行
func (s *dashboardService) GetUserTrafficRanking(limit int) ([]UserTrafficRank, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	db := database.Get()
	var rankings []UserTrafficRank

	// 查询用户流量消耗排行
	err := db.Model(&model.User{}).
		Select("id as user_id, email, used_traffic as total, 0 as upload, 0 as download").
		Where("status = ?", 1).
		Order("used_traffic DESC").
		Limit(limit).
		Scan(&rankings).Error

	if err != nil {
		return nil, err
	}

	return rankings, nil
}

// GetInviteRanking 获取邀请排行
func (s *dashboardService) GetInviteRanking(limit int) ([]InviteRank, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	db := database.Get()
	var rankings []InviteRank

	// 查询邀请排行（基于邀请码使用记录）
	err := db.Model(&model.InviteCode{}).
		Select("user_id, users.email, COUNT(*) as invite_count, COALESCE(SUM(commission_logs.amount), 0) as commission").
		Joins("LEFT JOIN users ON users.id = invite_codes.user_id").
		Joins("LEFT JOIN commission_logs ON commission_logs.user_id = invite_codes.user_id").
		Where("invite_codes.status = ?", 1). // 已使用的邀请码
		Group("invite_codes.user_id, users.email").
		Order("invite_count DESC").
		Limit(limit).
		Scan(&rankings).Error

	if err != nil {
		// 如果查询失败，返回空列表
		return []InviteRank{}, nil
	}

	return rankings, nil
}

// GetCommissionStats 获取佣金统计
func (s *dashboardService) GetCommissionStats() (*CommissionStats, error) {
	db := database.Get()
	stats := &CommissionStats{}

	// 总佣金
	if err := db.Model(&model.CommissionLog{}).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalCommission).Error; err != nil {
		return nil, err
	}

	// 已结算佣金
	if err := db.Model(&model.CommissionLog{}).
		Where("status = ?", 1).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.SettledCommission).Error; err != nil {
		return nil, err
	}

	// 待结算佣金
	if err := db.Model(&model.CommissionLog{}).
		Where("status = ?", 0).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.PendingCommission).Error; err != nil {
		return nil, err
	}

	// 总邀请人数
	if err := db.Model(&model.InviteCode{}).
		Where("status = ?", 1).
		Count(&stats.TotalInvites).Error; err != nil {
		stats.TotalInvites = 0
	}

	return stats, nil
}

// GetComprehensiveStats 获取综合统计
func (s *dashboardService) GetComprehensiveStats() (*ComprehensiveStats, error) {
	db := database.Get()
	stats := &ComprehensiveStats{}

	now := time.Now()
	todayStart := now.Truncate(24 * time.Hour)
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonthStart := currentMonthStart.AddDate(0, -1, 0)
	twoMonthsAgoStart := currentMonthStart.AddDate(0, -2, 0)

	// 收入相关
	db.Model(&model.Order{}).
		Where("created_at >= ? AND status = ?", todayStart, model.OrderStatusPaid).
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TodayIncome)

	var yesterdayIncome float64
	db.Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ? AND status = ?", yesterdayStart, todayStart, model.OrderStatusPaid).
		Select("COALESCE(SUM(amount), 0)").Scan(&yesterdayIncome)

	db.Model(&model.Order{}).
		Where("created_at >= ? AND status = ?", currentMonthStart, model.OrderStatusPaid).
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.CurrentMonthIncome)

	db.Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ? AND status = ?", lastMonthStart, currentMonthStart, model.OrderStatusPaid).
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.LastMonthIncome)

	var twoMonthsAgoIncome float64
	db.Model(&model.Order{}).
		Where("created_at >= ? AND created_at < ? AND status = ?", twoMonthsAgoStart, lastMonthStart, model.OrderStatusPaid).
		Select("COALESCE(SUM(amount), 0)").Scan(&twoMonthsAgoIncome)

	// 增长率计算
	if yesterdayIncome > 0 {
		stats.DayIncomeGrowth = (stats.TodayIncome - yesterdayIncome) / yesterdayIncome * 100
	}
	if stats.LastMonthIncome > 0 {
		stats.MonthIncomeGrowth = (stats.CurrentMonthIncome - stats.LastMonthIncome) / stats.LastMonthIncome * 100
	}

	// 佣金相关
	db.Model(&model.CommissionLog{}).
		Where("created_at >= ?", currentMonthStart).
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.CurrentMonthCommissionPayout)

	db.Model(&model.CommissionLog{}).
		Where("created_at >= ? AND created_at < ?", lastMonthStart, currentMonthStart).
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.LastMonthCommissionPayout)

	var twoMonthsAgoCommission float64
	db.Model(&model.CommissionLog{}).
		Where("created_at >= ? AND created_at < ?", twoMonthsAgoStart, lastMonthStart).
		Select("COALESCE(SUM(amount), 0)").Scan(&twoMonthsAgoCommission)

	if twoMonthsAgoCommission > 0 {
		stats.CommissionGrowth = (stats.LastMonthCommissionPayout - twoMonthsAgoCommission) / twoMonthsAgoCommission * 100
	}

	db.Model(&model.Order{}).
		Where("commission_status = 0 AND invite_user_id IS NOT NULL AND status = ? AND commission_balance > 0", model.OrderStatusPaid).
		Count(&stats.CommissionPendingTotal)

	// 用户相关
	db.Model(&model.User{}).Where("created_at >= ?", currentMonthStart).Count(&stats.CurrentMonthNewUsers)
	db.Model(&model.User{}).Count(&stats.TotalUsers)
	db.Model(&model.User{}).Where("expired_at IS NULL OR expired_at > ?", now).Count(&stats.ActiveUsers)

	var lastMonthNewUsers int64
	db.Model(&model.User{}).Where("created_at >= ? AND created_at < ?", lastMonthStart, currentMonthStart).Count(&lastMonthNewUsers)
	if lastMonthNewUsers > 0 {
		stats.UserGrowth = float64(stats.CurrentMonthNewUsers-lastMonthNewUsers) / float64(lastMonthNewUsers) * 100
	}

	db.Model(&model.User{}).Where("online_count > 0").Count(&stats.OnlineUsers)
	db.Model(&model.User{}).Select("COALESCE(SUM(online_count), 0)").Scan(&stats.OnlineDevices)

	// 工单相关
	db.Model(&model.Ticket{}).Where("status = ?", 0).Count(&stats.TicketPendingTotal)

	// 节点相关
	db.Model(&model.Node{}).Where("status = 1").Count(&stats.OnlineNodes)

	// 流量统计
	type trafficResult struct {
		Upload   int64
		Download int64
	}
	var todayTraffic, monthTraffic, totalTraffic trafficResult

	db.Model(&model.TrafficLog{}).Where("recorded_at >= ?", todayStart).
		Select("COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download").
		Scan(&todayTraffic)
	stats.TodayTraffic = TrafficSummary{Upload: todayTraffic.Upload, Download: todayTraffic.Download, Total: todayTraffic.Upload + todayTraffic.Download}

	db.Model(&model.TrafficLog{}).Where("recorded_at >= ?", currentMonthStart).
		Select("COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download").
		Scan(&monthTraffic)
	stats.MonthTraffic = TrafficSummary{Upload: monthTraffic.Upload, Download: monthTraffic.Download, Total: monthTraffic.Upload + monthTraffic.Download}

	db.Model(&model.TrafficLog{}).
		Select("COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download").
		Scan(&totalTraffic)
	stats.TotalTraffic = TrafficSummary{Upload: totalTraffic.Upload, Download: totalTraffic.Download, Total: totalTraffic.Upload + totalTraffic.Download}

	return stats, nil
}
