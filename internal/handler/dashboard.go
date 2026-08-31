package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// DashboardHandler 看板处理器
type DashboardHandler struct {
	dashboardService service.DashboardService
}

// NewDashboardHandler 创建看板处理器
func NewDashboardHandler(dashboardService service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetOverview 获取概览数据
// @Summary 数据概览
// @Description 获取管理员看板概览数据（用户/订单/收入/节点等）
// @Tags 管理员-看板
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=service.DashboardOverview}
// @Router /api/v1/admin/dashboard/overview [get]
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	overview, err := h.dashboardService.GetOverview()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, overview)
}

// GetRecentOrders 最近订单
// @Summary 最近订单
// @Description 获取最近的订单列表
// @Tags 管理员-看板
// @Produce json
// @Security Bearer
// @Param limit query int false "数量" default(10)
// @Success 200 {object} response.Response{data=[]model.Order}
// @Router /api/v1/admin/dashboard/recent-orders [get]
func (h *DashboardHandler) GetRecentOrders(c *gin.Context) {
	limit := atoi2(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	orders, err := h.dashboardService.GetRecentOrders(limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, orders)
}

// GetRecentUsers 最近注册用户
// @Summary 最近注册用户
// @Description 获取最近注册的用户列表
// @Tags 管理员-看板
// @Produce json
// @Security Bearer
// @Param limit query int false "数量" default(10)
// @Success 200 {object} response.Response{data=[]model.User}
// @Router /api/v1/admin/dashboard/recent-users [get]
func (h *DashboardHandler) GetRecentUsers(c *gin.Context) {
	limit := atoi2(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, err := h.dashboardService.GetRecentUsers(limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, users)
}

// GetIncomeStats 收入统计
// @Summary 收入统计
// @Description 获取最近N天的收入趋势
// @Tags 管理员-看板
// @Produce json
// @Security Bearer
// @Param days query int false "天数" default(7)
// @Success 200 {object} response.Response{data=[]service.DailyIncome}
// @Router /api/v1/admin/dashboard/income-stats [get]
func (h *DashboardHandler) GetIncomeStats(c *gin.Context) {
	days := atoi2(c.DefaultQuery("days", "7"))
	if days < 1 || days > 365 {
		days = 7
	}

	stats, err := h.dashboardService.GetIncomeStats(days)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

// GetUserGrowthStats 用户增长统计
// @Summary 用户增长统计
// @Description 获取最近N天的用户注册趋势
// @Tags 管理员-看板
// @Produce json
// @Security Bearer
// @Param days query int false "天数" default(7)
// @Success 200 {object} response.Response{data=[]service.DailyUserGrowth}
// @Router /api/v1/admin/dashboard/user-growth [get]
func (h *DashboardHandler) GetUserGrowthStats(c *gin.Context) {
	days := atoi2(c.DefaultQuery("days", "7"))
	if days < 1 || days > 365 {
		days = 7
	}

	stats, err := h.dashboardService.GetUserGrowthStats(days)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

// GetNodeStats 节点统计
// @Summary 节点统计
// @Description 获取节点在线/离线统计
// @Tags 管理员-看板
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=service.NodeStats}
// @Router /api/v1/admin/dashboard/node-stats [get]
func (h *DashboardHandler) GetNodeStats(c *gin.Context) {
	stats, err := h.dashboardService.GetNodeStats()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

// GetNodeTrafficRanking 节点流量排行
// @Summary 节点流量排行
// @Description 获取节点流量使用排行
// @Tags 管理员-看板
// @Produce json
// @Security Bearer
// @Param limit query int false "数量" default(10)
// @Success 200 {object} response.Response{data=[]service.NodeTrafficRank}
// @Router /api/v1/admin/dashboard/node-traffic-ranking [get]
func (h *DashboardHandler) GetNodeTrafficRanking(c *gin.Context) {
	limit := atoi2(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	rankings, err := h.dashboardService.GetNodeTrafficRanking(limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, rankings)
}

// GetUserTrafficRanking 用户流量排行
// @Summary 用户流量排行
// @Description 获取用户流量消耗排行
// @Tags 管理员-看板
// @Produce json
// @Security Bearer
// @Param limit query int false "数量" default(10)
// @Success 200 {object} response.Response{data=[]service.UserTrafficRank}
// @Router /api/v1/admin/dashboard/user-traffic-ranking [get]
func (h *DashboardHandler) GetUserTrafficRanking(c *gin.Context) {
	limit := atoi2(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	rankings, err := h.dashboardService.GetUserTrafficRanking(limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, rankings)
}

// GetInviteRanking 邀请排行
// @Summary 邀请排行
// @Description 获取用户邀请排行
// @Tags 管理员-看板
// @Produce json
// @Security Bearer
// @Param limit query int false "数量" default(10)
// @Success 200 {object} response.Response{data=[]service.InviteRank}
// @Router /api/v1/admin/dashboard/invite-ranking [get]
func (h *DashboardHandler) GetInviteRanking(c *gin.Context) {
	limit := atoi2(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	rankings, err := h.dashboardService.GetInviteRanking(limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, rankings)
}

// GetCommissionStats 佣金统计
// @Summary 佣金统计
// @Description 获取佣金统计数据
// @Tags 管理员-看板
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=service.CommissionStats}
// @Router /api/v1/admin/dashboard/commission-stats [get]
func (h *DashboardHandler) GetCommissionStats(c *gin.Context) {
	stats, err := h.dashboardService.GetCommissionStats()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

// GetComprehensiveStats 综合统计数据
// @Summary 综合统计
// @Description 获取包含收入、用户、流量等全面统计数据
// @Tags 管理员-看板
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/v1/admin/dashboard/comprehensive-stats [get]
func (h *DashboardHandler) GetComprehensiveStats(c *gin.Context) {
	stats, err := h.dashboardService.GetComprehensiveStats()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

func atoi2(s string) int {
	n := 0
	fmt.Sscanf(s, "%d", &n)
	return n
}
