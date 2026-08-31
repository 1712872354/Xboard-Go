package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"xboard-go/config"
	"xboard-go/internal/grpc"
	"xboard-go/internal/handler"
	"xboard-go/internal/model"
	"xboard-go/internal/repository"
	"xboard-go/internal/router"
	"xboard-go/internal/scheduler/tasks"
	"xboard-go/internal/service"
	"xboard-go/pkg/database"
	"xboard-go/pkg/logger"
	"xboard-go/pkg/redis"
	"xboard-go/pkg/scheduler"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	// 1. 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 检查JWT密钥是否为默认值
	if cfg.JWT.Secret == "xboard-go-secret-key-change-in-production" {
		fmt.Println("WARNING: Using default JWT secret key. Please change it in production!")
	}

	// 2. 初始化日志
	if err := logger.Init(&cfg.Log); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Sugar().Infof("Starting %s v%s...", cfg.App.Name, cfg.App.Version)

	// 3. SQLite 目录自动创建
	if cfg.Database.Driver == "sqlite" || cfg.Database.Driver == "sqlite3" {
		dbDir := filepath.Dir(cfg.Database.DBName)
		if dbDir != "" && dbDir != "." {
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				logger.Sugar().Fatalf("Failed to create database directory %s: %v", dbDir, err)
			}
		}
	}

	// 4. 初始化数据库
	if err := database.Init(&cfg.Database); err != nil {
		logger.Sugar().Fatalf("Failed to init database: %v", err)
	}

	// 5. 数据库自动迁移
	if err := database.AutoMigrate(
		&model.User{},
		&model.Plan{},
		&model.Order{},
		&model.Node{},
		&model.NodeUser{},
		&model.TrafficLog{},
		&model.RedeemCode{},
		&model.Ticket{},
		&model.TicketReply{},
		&model.Setting{},
		&model.Notice{},
		&model.Knowledge{},
		&model.MailTemplate{},
		&model.MailLog{},
		&model.Coupon{},
		&model.GiftCardTemplate{},
		&model.GiftCardCode{},
		&model.GiftCardUsage{},
		&model.InviteCode{},
		&model.CommissionLog{},
		&model.ServerGroup{},
		&model.ServerRoute{},
		&model.ServerMachine{},
		&model.ServerMachineLoadHistory{},
		&model.ServerStat{},
		&model.ServerLog{},
		&model.Plugin{},
		&model.SubscribeTemplate{},
		&model.AdminAuditLog{},
		&model.NodeTemplate{},
		&model.Payment{},
		&model.TrafficResetLog{},
		&model.Stat{},
		&model.StatServer{},
		&model.StatUser{},
	); err != nil {
		logger.Sugar().Fatalf("Failed to auto migrate: %v", err)
	}
	logger.Sugar().Info("Database migration completed")

	// Wire the metrics cache from gRPC package to HTTP handler.
	handler.GlobalNodeMetricsFunc = func(nodeID uint32) interface{} {
		return grpc.GlobalMetricsCache.GetMetrics(nodeID)
	}

	// 6. 初始化 Redis
	if err := redis.Init(&cfg.Redis); err != nil {
		logger.Sugar().Warnf("Failed to init redis: %v (continuing without redis)", err)
	}

	// 7. 提前初始化 gRPC Broadcaster（确保 REST/WS 节点端点可用）
	// 如果 gRPC 未启用，也创建一个空的 Broadcaster 供 REST/WS 使用
	if grpc.NodeBroadcaster == nil {
		grpc.NodeBroadcaster = grpc.NewBroadcaster()
	}

	// 8. 设置路由（含前端 SPA 静态文件服务）
	r := router.SetupRouter(cfg)

	// 9. 启动 gRPC 服务（可选）
	var grpcSrv *grpc.GRPCServer
	if cfg.GRPC.Enabled {
		grpcPort := cfg.GRPC.Port
		if grpcPort == 0 {
			grpcPort = 50051
		}
		grpcSrv = grpc.NewServer(grpcPort, cfg)

		// Wire HTTP handlers to gRPC broadcaster for real-time push.
		handler.NotifyNodeConfigChange = func(nodeID uint, name, typ, address string, port int, serverInfo string, rate float64, groupID, parentID uint) {
			if grpc.NodeBroadcaster != nil && grpc.NodeBroadcaster.HasSubscriber(uint32(nodeID)) {
				grpc.NodeBroadcaster.BroadcastConfig(uint32(nodeID), &grpc.NodeConfig{
					ID:         uint32(nodeID),
					Name:       name,
					Protocol:   typ,
					Address:    address,
					Port:       int32(port),
					ServerInfo: serverInfo,
					Rate:       float32(rate),
					GroupID:    uint32(groupID),
					ParentID:   uint32(parentID),
				})
				logger.Sugar().Infof("Broadcasted config update to node %d", nodeID)
			}
		}
		handler.NotifyUserChange = func() {
			// Broadcast full user list to all connected nodes.
			if grpc.NodeBroadcaster != nil {
				users, err := grpc.GetActiveUsersForBroadcast()
				if err != nil {
					logger.Sugar().Warnf("Failed to get users for broadcast: %v", err)
					return
				}
				for _, nodeID := range grpc.NodeBroadcaster.GetSubscriberNodeIDs() {
					grpc.NodeBroadcaster.BroadcastUsers(nodeID, users)
				}
				logger.Sugar().Infof("Broadcasted user list update to %d nodes", grpc.NodeBroadcaster.SubscriberCount())
			}
		}

		go func() {
			if err := grpcSrv.Start(); err != nil {
				logger.Sugar().Fatalf("gRPC server error: %v", err)
			}
		}()
	}

	// 10. 启动节点健康检查服务
	healthCtx, healthCancel := context.WithCancel(context.Background())
	defer healthCancel()
	{
		nodeRepo := repository.NewNodeRepository()
		machineRepo := repository.NewServerMachineRepository()
		healthSvc := service.NewNodeHealthService(nodeRepo, machineRepo)
		go healthSvc.Start(healthCtx)
		logger.Sugar().Info("Node health check service started")
	}

	// 11. 启动定时任务调度器
	{
		sched := scheduler.NewScheduler(logger.Get())
		zapLogger := logger.Get()

		orderTasks := tasks.NewOrderTasks(database.Get(), zapLogger)
		commissionTasks := tasks.NewCommissionTasks(database.Get(), zapLogger)
		trafficTasks := tasks.NewTrafficTasks(database.Get(), zapLogger)
		trafficResetTasks := tasks.NewTrafficResetTasks(database.Get(), zapLogger)
		mailTasks := tasks.NewMailTasks(database.Get(), zapLogger)
		statTasks := tasks.NewStatTasks(database.Get(), zapLogger)
		ticketTasks := tasks.NewTicketTasks(database.Get(), zapLogger)
		cleanupTasks := tasks.NewCleanupTasks(database.Get(), zapLogger)

		tasksToRegister := []scheduler.Task{
			// 每分钟任务
			{Name: "check_pending_orders", Spec: "0 */1 * * * *", Func: orderTasks.CheckPendingOrders, Enabled: true},
			{Name: "check_commission", Spec: "0 */1 * * * *", Func: commissionTasks.CheckCommission, Enabled: true},
			{Name: "check_traffic_exceeded", Spec: "0 */1 * * * *", Func: trafficTasks.CheckTrafficExceeded, Enabled: true},
			{Name: "check_and_reset_traffic", Spec: "0 */1 * * * *", Func: trafficResetTasks.CheckAndResetTraffic, Enabled: true},
			{Name: "check_pending_tickets", Spec: "0 */1 * * * *", Func: ticketTasks.CheckPendingTickets, Enabled: true},
			// 每5分钟任务
			{Name: "check_processing_orders", Spec: "0 */5 * * * *", Func: orderTasks.CheckProcessingOrders, Enabled: true},
			{Name: "check_stale_replies", Spec: "0 */5 * * * *", Func: ticketTasks.CheckStaleReplies, Enabled: true},
			{Name: "clean_online_status", Spec: "0 */5 * * * *", Func: cleanupTasks.CleanOnlineStatus, Enabled: true},
			// 每日任务
			{Name: "generate_daily_stats", Spec: "0 10 0 * * *", Func: statTasks.GenerateDailyStats, Enabled: true},
			{Name: "generate_daily_server_stats", Spec: "0 15 0 * * *", Func: statTasks.GenerateDailyServerStats, Enabled: true},
			{Name: "generate_daily_user_stats", Spec: "0 20 0 * * *", Func: statTasks.GenerateDailyUserStats, Enabled: true},
			{Name: "send_remind_mails", Spec: "0 30 11 * * *", Func: mailTasks.SendRemindMails, Enabled: true},
			{Name: "set_initial_reset_times", Spec: "0 0 3 * * *", Func: trafficResetTasks.SetInitialResetTimes, Enabled: true},
			{Name: "clean_expired_tickets", Spec: "0 0 2 * * *", Func: cleanupTasks.CleanExpiredTickets, Enabled: true},
		}

		for _, task := range tasksToRegister {
			if err := sched.Register(task); err != nil {
				logger.Sugar().Warnf("Failed to register task %s: %v", task.Name, err)
			}
		}

		if err := sched.Start(); err != nil {
			logger.Sugar().Warnf("Failed to start scheduler: %v", err)
		} else {
			logger.Sugar().Info("Scheduler started with tasks")
		}
	}

	// 12. 启动 HTTP 服务
	addr := cfg.Server.Addr()
	logger.Sugar().Infof("HTTP server listening on %s", addr)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := r.Run(addr); err != nil {
			logger.Sugar().Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	logger.Sugar().Info("Server shutting down...")

	// 停止健康检查服务
	healthCancel()

	// 停止 gRPC 服务
	if grpcSrv != nil {
		grpcSrv.Stop()
	}

	logger.Sugar().Info("Server exited")
}
