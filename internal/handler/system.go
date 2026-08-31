package handler

import (
	"runtime"
	"time"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"
	appredis "xboard-go/pkg/redis"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SystemHandler 系统管理处理器
type SystemHandler struct {
	db *gorm.DB
}

// NewSystemHandler 创建系统处理器
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{
		db: database.Get(),
	}
}

// GetSystemStatus 获取系统状态
func (h *SystemHandler) GetSystemStatus(c *gin.Context) {
	status := gin.H{
		"server":  "running",
		"version": getVersion(),
		"uptime":  getUptime(),
		"go_version": runtime.Version(),
	}

	// 检查数据库连接
	sqlDB, err := h.db.DB()
	if err == nil {
		if err := sqlDB.Ping(); err != nil {
			status["database"] = "error"
		} else {
			status["database"] = "connected"
		}
	}

	// 检查 Redis 连接
	redisClient := appredis.Client()
	if redisClient != nil {
		ctx := c.Request.Context()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			status["redis"] = "error"
		} else {
			status["redis"] = "connected"
		}
	} else {
		status["redis"] = "not_configured"
	}

	// 获取在线用户数
	var onlineUsers int64
	h.db.Model(&model.User{}).Where("online_count > 0").Count(&onlineUsers)
	status["online_users"] = onlineUsers

	// 获取在线节点数
	var onlineNodes int64
	h.db.Model(&model.Node{}).Where("status = 1").Count(&onlineNodes)
	status["online_nodes"] = onlineNodes

	// 获取待处理工单数
	var pendingTickets int64
	h.db.Model(&model.Ticket{}).Where("status = ?", model.TicketStatusOpen).Count(&pendingTickets)
	status["pending_tickets"] = pendingTickets

	// 获取今日注册数
	today := time.Now().Truncate(24 * time.Hour)
	var todayRegisters int64
	h.db.Model(&model.User{}).Where("created_at >= ?", today).Count(&todayRegisters)
	status["today_registers"] = todayRegisters

	// 获取今日订单数
	var todayOrders int64
	h.db.Model(&model.Order{}).Where("created_at >= ?", today).Count(&todayOrders)
	status["today_orders"] = todayOrders

	response.Success(c, status)
}

// GetSystemInfo 获取系统信息
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	info := gin.H{
		"version":      getVersion(),
		"go_version":   runtime.Version(),
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
		"num_cpu":      runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
		"memory": gin.H{
			"alloc":       m.Alloc,
			"total_alloc": m.TotalAlloc,
			"sys":         m.Sys,
			"num_gc":      m.NumGC,
		},
		"uptime": getUptime(),
	}

	response.Success(c, info)
}

// GetSchedulerStatus 获取定时任务状态
func (h *SystemHandler) GetSchedulerStatus(c *gin.Context) {
	// 从设置获取最后运行时间
	var lastRunAt string
	var setting model.Setting
	if err := h.db.Where("`key` = ?", "scheduler_last_run_at").First(&setting).Error; err == nil {
		lastRunAt = setting.Value
	}

	// 检查定时任务是否健康（最近2分钟内有运行）
	healthy := false
	if lastRunAt != "" {
		if t, err := time.Parse(time.RFC3339, lastRunAt); err == nil {
			healthy = time.Since(t) < 2*time.Minute
		}
	}

	status := gin.H{
		"healthy":     healthy,
		"last_run_at": lastRunAt,
		"tasks": []gin.H{
			{"name": "check_pending_orders", "schedule": "*/1 * * * *"},
			{"name": "check_commission", "schedule": "*/1 * * * *"},
			{"name": "check_traffic_exceeded", "schedule": "*/1 * * * *"},
			{"name": "check_and_reset_traffic", "schedule": "*/1 * * * *"},
			{"name": "check_pending_tickets", "schedule": "*/1 * * * *"},
			{"name": "generate_daily_stats", "schedule": "0 10 0 * * *"},
			{"name": "send_remind_mails", "schedule": "30 11 * * *"},
		},
	}

	response.Success(c, status)
}

// 版本和启动时间
var (
	startTime = time.Now()
	version   = "dev"
)

func getVersion() string {
	return version
}

func getUptime() string {
	return time.Since(startTime).Round(time.Second).String()
}
