package router

import (
	"time"

	"xboard-go/config"
	"xboard-go/internal/grpc"
	"xboard-go/internal/handler"
	"xboard-go/internal/middleware"
	"xboard-go/internal/repository"
	"xboard-go/internal/service"
	"xboard-go/internal/static"
	"xboard-go/pkg/ratelimit"
	appredis "xboard-go/pkg/redis"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置路由
func SetupRouter(cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recovery())
	r.Use(middleware.SafeModeMiddleware())

	// 限流中间件（可选启用）
	if cfg.RateLimit.Enabled {
		var limiter ratelimit.Limiter

		// 优先使用 Redis 限流器（支持分布式）
		redisClient := appredis.Client()
		if redisClient != nil {
			limiter = ratelimit.NewRedisLimiter(redisClient, "xboard:rate_limit")
		} else {
			// Redis 不可用时降级到内存限流器（单节点）
			limiter = ratelimit.NewMemoryLimiter("xboard:rate_limit")
		}

		rlConfig := middleware.RateLimitConfig{
			IPLimit:       cfg.RateLimit.IPLimit,
			IPWindow:      time.Minute,
			UserLimit:     cfg.RateLimit.UserLimit,
			UserWindow:    time.Minute,
			IPWhitelist:   cfg.RateLimit.IPWhitelist,
			PathWhitelist: cfg.RateLimit.PathWhitelist,
		}

		r.Use(middleware.RateLimitMiddleware(limiter, rlConfig))
	}

	// 健康检查
	healthHandler := handler.NewHealthHandler()
	r.GET("/healthz", healthHandler.Check)

	// 初始化仓储层
	userRepo := repository.NewUserRepository()
	planRepo := repository.NewPlanRepository()
	orderRepo := repository.NewOrderRepository()
	nodeRepo := repository.NewNodeRepository()
	redeemCodeRepo := repository.NewRedeemCodeRepository()
	ticketRepo := repository.NewTicketRepository()
	settingRepo := repository.NewSettingRepository()
	noticeRepo := repository.NewNoticeRepository()
	knowledgeRepo := repository.NewKnowledgeRepository()
	mailTemplateRepo := repository.NewMailTemplateRepository()
	couponRepo := repository.NewCouponRepository()
	giftCardTemplateRepo := repository.NewGiftCardTemplateRepository()
	giftCardCodeRepo := repository.NewGiftCardCodeRepository()
	giftCardUsageRepo := repository.NewGiftCardUsageRepository()
	inviteCodeRepo := repository.NewInviteCodeRepository()
	commissionLogRepo := repository.NewCommissionLogRepository()
	serverGroupRepo := repository.NewServerGroupRepository()
	serverRouteRepo := repository.NewServerRouteRepository()
	serverMachineRepo := repository.NewServerMachineRepository()
	pluginRepo := repository.NewPluginRepository()
	auditLogRepo := repository.NewAdminAuditLogRepository()
	nodeTemplateRepo := repository.NewNodeTemplateRepository()
	paymentRepo := repository.NewPaymentRepository()

	// 初始化服务层
	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)
	planService := service.NewPlanService(planRepo)
	orderService := service.NewOrderService(orderRepo, planRepo, userRepo, couponRepo)
	subscribeService := service.NewSubscribeService(userRepo, nodeRepo, planRepo)
	nodeService := service.NewNodeService(nodeRepo)
	trafficService := service.NewTrafficService()
	_ = trafficService // used by gRPC server

	// 节点通讯服务（REST API + WebSocket 用于 Xboard-Node）
	nodeUserRepo := repository.NewNodeUserRepository()
	trafficLogRepo := repository.NewTrafficLogRepository()
	uniProxySvc := service.NewUniProxyService(nodeRepo, nodeUserRepo, userRepo, trafficLogRepo, trafficService)

	// 支付服务（使用服务地址构建回调URL）
	baseURL := "http://" + cfg.Server.Addr()
	paymentService := service.NewPaymentService(orderRepo, userService, baseURL)
	redeemService := service.NewRedeemService(redeemCodeRepo, planRepo, userService)
	ticketService := service.NewTicketService(ticketRepo, userRepo)
	dashboardService := service.NewDashboardService(userRepo, orderRepo, planRepo, nodeRepo, ticketRepo, redeemCodeRepo)
	settingService := service.NewSettingService(settingRepo)
	noticeService := service.NewNoticeService(noticeRepo)
	knowledgeService := service.NewKnowledgeService(knowledgeRepo)
	mailTemplateService := service.NewMailTemplateService(mailTemplateRepo)
	couponService := service.NewCouponService(couponRepo)
	giftCardService := service.NewGiftCardService(giftCardTemplateRepo, giftCardCodeRepo, giftCardUsageRepo)
	inviteService := service.NewInviteService(inviteCodeRepo, commissionLogRepo, userRepo)
	serverGroupService := service.NewServerGroupService(serverGroupRepo)
	serverRouteService := service.NewServerRouteService(serverRouteRepo)
	serverMachineService := service.NewServerMachineService(serverMachineRepo)
	pluginService := service.NewPluginService(pluginRepo)
	auditLogService := service.NewAdminAuditLogService(auditLogRepo)
	nodeTemplateService := service.NewNodeTemplateService(nodeTemplateRepo)
	deviceService := service.NewDeviceService(userRepo, nodeRepo)
	paymentGatewayService := service.NewPaymentGatewayService(paymentRepo)
	telegramService := service.NewTelegramService()

	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	planHandler := handler.NewPlanHandler(planService)
	orderHandler := handler.NewOrderHandler(orderService)
	subscribeHandler := handler.NewSubscribeHandler(subscribeService)
	nodeHandler := handler.NewNodeHandler(nodeService)
	trafficHandler := handler.NewTrafficHandler(trafficService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	redeemHandler := handler.NewRedeemHandler(redeemService)
	ticketHandler := handler.NewTicketHandler(ticketService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	settingHandler := handler.NewSettingHandler(settingService)
	noticeHandler := handler.NewNoticeHandler(noticeService)
	knowledgeHandler := handler.NewKnowledgeHandler(knowledgeService)
	mailTemplateHandler := handler.NewMailTemplateHandler(mailTemplateService)
	couponHandler := handler.NewCouponHandler(couponService)
	giftCardHandler := handler.NewGiftCardHandler(giftCardService)
	inviteHandler := handler.NewInviteHandler(inviteService)
	serverGroupHandler := handler.NewServerGroupHandler(serverGroupService)
	serverRouteHandler := handler.NewServerRouteHandler(serverRouteService)
	serverMachineHandler := handler.NewServerMachineHandler(serverMachineService)
	pluginHandler := handler.NewPluginHandler(pluginService)
	auditLogHandler := handler.NewAdminAuditLogHandler(auditLogService)
	nodeTemplateHandler := handler.NewNodeTemplateHandler(nodeTemplateService)
	deviceHandler := handler.NewDeviceHandler(deviceService)
	paymentGatewayHandler := handler.NewPaymentGatewayHandler(paymentGatewayService)
	telegramHandler := handler.NewTelegramHandler(telegramService)
	systemHandler := handler.NewSystemHandler()
	updateHandler := handler.NewUpdateHandler(cfg.App.Version)
	commHandler := handler.NewCommHandler()

	// 节点 REST API 处理器（兼容 Xboard-Node）
	nodeServerHandler := handler.NewNodeServerHandler(uniProxySvc)
	// 节点 WebSocket 处理器（兼容 Xboard-Node 实时推送）
	nodeWSHandler := handler.NewNodeWSHandler(grpc.NodeBroadcaster)

	// API v1 路由组
	apiV1 := r.Group("/api/v1")
	{
		// 公开路由（不需要认证）
		public := apiV1.Group("")
		{
			// 认证相关
			auth := public.Group("/auth")
			{
				auth.POST("/register", middleware.CaptchaMiddleware(), middleware.RegisterRateLimitMiddleware(), authHandler.Register)
				auth.POST("/login", middleware.CaptchaMiddleware(), middleware.LoginSecurityMiddleware(), authHandler.Login)
				auth.POST("/refresh", authHandler.RefreshToken)
				auth.POST("/forget", authHandler.ForgetPassword)
				auth.POST("/reset", authHandler.ResetPassword)
				auth.POST("/send-code", authHandler.SendVerificationCode)
				auth.POST("/verify", authHandler.VerifyEmail)
			}

			// 套餐列表（公开）
			public.GET("/plans", planHandler.ListPlans)
			public.GET("/plans/:id", planHandler.GetPlan)

			// 客户端订阅接口（公开，使用token验证）
			public.GET("/client/subscribe", subscribeHandler.ClientSubscribe)

			// Telegram Webhook（公开）
			public.POST("/telegram/webhook", telegramHandler.Webhook)
		}

		// 需要认证的用户路由
		user := apiV1.Group("/user")
		user.Use(middleware.JWTAuth())
		{
			user.GET("/profile", userHandler.GetProfile)
			user.PUT("/profile", userHandler.UpdateProfile)
			user.PUT("/password", userHandler.ChangePassword)
			user.POST("/subscribe/reset", userHandler.ResetSubscribeToken)

			// 流量相关
			user.GET("/traffic/stats", trafficHandler.GetStats)
			user.GET("/traffic/history", trafficHandler.GetHistory)
			user.GET("/traffic/daily", trafficHandler.GetDailyTraffic)
		}

		// 订单路由（需要认证）
		orders := apiV1.Group("/orders")
		orders.Use(middleware.JWTAuth())
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("", orderHandler.ListOrders)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.POST("/:id/cancel", orderHandler.CancelOrder)
		}

		// 支付路由
		payment := apiV1.Group("/payment")
		{
			// 获取支付方式列表（公开）
			payment.GET("/methods", paymentHandler.ListMethods)

			// 模拟支付页面（公开，用于测试）
			payment.GET("/mock/pay", paymentHandler.MockPay)
			payment.GET("/mock/callback", paymentHandler.MockCallback)

			// 需要认证的支付接口
			paymentAuth := payment.Group("")
			paymentAuth.Use(middleware.JWTAuth())
			{
				paymentAuth.POST("/create", paymentHandler.CreatePayment)
			}
		}

		// 卡密兑换（需要认证）
		redeem := apiV1.Group("/redeem")
		redeem.Use(middleware.JWTAuth())
		{
			redeem.POST("", redeemHandler.Redeem)
		}

		// 工单路由（需要认证）
		tickets := apiV1.Group("/tickets")
		tickets.Use(middleware.JWTAuth())
		{
			tickets.GET("", ticketHandler.ListUserTickets)
			tickets.POST("", ticketHandler.CreateTicket)
			tickets.GET("/stats", ticketHandler.GetUserStats)
			tickets.GET("/:id", ticketHandler.GetTicket)
			tickets.POST("/:id/reply", ticketHandler.Reply)
			tickets.POST("/:id/close", ticketHandler.CloseTicket)
		}

		// 管理员路由
		admin := apiV1.Group("/admin")
		admin.Use(middleware.JWTAuth(), middleware.AdminRequired())
		{
			// 用户管理
			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("", userHandler.ListUsers)
				adminUsers.GET("/:id", userHandler.GetUser)
				adminUsers.PUT("/:id", userHandler.AdminUpdateUser)
				adminUsers.PUT("/:id/status", userHandler.UpdateUserStatus)
				adminUsers.PUT("/:id/role", userHandler.UpdateUserRole)
				adminUsers.DELETE("/:id", userHandler.DeleteUser)
				adminUsers.POST("/generate", userHandler.GenerateUsers)
				adminUsers.GET("/export", userHandler.ExportUsersCSV)
			}

			// 套餐管理
			adminPlans := admin.Group("/plans")
			{
				adminPlans.GET("", planHandler.ListAllPlans)
				adminPlans.POST("", planHandler.CreatePlan)
				adminPlans.PUT("/:id", planHandler.UpdatePlan)
				adminPlans.DELETE("/:id", planHandler.DeletePlan)
			}

			// 订单管理
			adminOrders := admin.Group("/orders")
			{
				adminOrders.GET("", orderHandler.ListAllOrders)
				adminOrders.POST("/confirm-payment", orderHandler.ConfirmPayment)
			}

			// 节点管理
			adminNodes := admin.Group("/nodes")
			{
				adminNodes.GET("", nodeHandler.ListNodes)
				adminNodes.POST("", nodeHandler.CreateNode)
				adminNodes.GET("/:id", nodeHandler.GetNode)
				adminNodes.PUT("/:id", nodeHandler.UpdateNode)
				adminNodes.DELETE("/:id", nodeHandler.DeleteNode)
				adminNodes.POST("/:id/copy", nodeHandler.CopyNode)
				adminNodes.PUT("/:id/status", nodeHandler.UpdateNodeStatus)
				adminNodes.GET("/:id/metrics", nodeHandler.GetNodeMetrics)
				adminNodes.POST("/:id/reset-traffic", nodeHandler.ResetNodeTraffic)
				adminNodes.POST("/batch", nodeHandler.BatchNodes)
				adminNodes.POST("/batch-reset-traffic", nodeHandler.BatchResetNodeTraffic)
			}

			// 节点模板管理
			adminNodeTemplates := admin.Group("/node-templates")
			{
				adminNodeTemplates.GET("", nodeTemplateHandler.ListNodeTemplates)
				adminNodeTemplates.POST("", nodeTemplateHandler.CreateNodeTemplate)
				adminNodeTemplates.GET("/:id", nodeTemplateHandler.GetNodeTemplate)
				adminNodeTemplates.PUT("/:id", nodeTemplateHandler.UpdateNodeTemplate)
				adminNodeTemplates.DELETE("/:id", nodeTemplateHandler.DeleteNodeTemplate)
			}

			// 流量管理
			adminTraffic := admin.Group("/traffic")
			{
				adminTraffic.POST("/sync", trafficHandler.SyncTraffic)
			}

			// 设备状态管理
			adminDevices := admin.Group("/devices")
			{
				adminDevices.GET("/user/:id", deviceHandler.GetUserDevices)
				adminDevices.GET("/node/:id", deviceHandler.GetNodeDevices)
				adminDevices.GET("/user/:id/count", deviceHandler.GetOnlineDeviceCount)
				adminDevices.POST("/cleanup", deviceHandler.CleanupOfflineDevices)
			}

			// 管理员查看用户流量
			adminUsers.GET("/:id/traffic", trafficHandler.AdminGetUserTraffic)

			// 卡密管理
			adminRedeem := admin.Group("/redeem")
			{
				adminRedeem.GET("", redeemHandler.List)
				adminRedeem.POST("/generate", redeemHandler.Generate)
				adminRedeem.DELETE("/:id", redeemHandler.Delete)
				adminRedeem.GET("/stats", redeemHandler.GetStats)
			}

			// 工单管理
			adminTickets := admin.Group("/tickets")
			{
				adminTickets.GET("", ticketHandler.ListAllTickets)
				adminTickets.GET("/:id", ticketHandler.GetTicket)
				adminTickets.POST("/:id/reply", ticketHandler.Reply)
				adminTickets.POST("/:id/close", ticketHandler.CloseTicket)
				adminTickets.DELETE("/:id", ticketHandler.DeleteTicket)
			}

			// 数据看板
			adminDashboard := admin.Group("/dashboard")
			{
				adminDashboard.GET("/overview", dashboardHandler.GetOverview)
				adminDashboard.GET("/recent-orders", dashboardHandler.GetRecentOrders)
				adminDashboard.GET("/recent-users", dashboardHandler.GetRecentUsers)
				adminDashboard.GET("/income-stats", dashboardHandler.GetIncomeStats)
				adminDashboard.GET("/user-growth", dashboardHandler.GetUserGrowthStats)
				adminDashboard.GET("/node-stats", dashboardHandler.GetNodeStats)
				adminDashboard.GET("/node-traffic-ranking", dashboardHandler.GetNodeTrafficRanking)
				adminDashboard.GET("/user-traffic-ranking", dashboardHandler.GetUserTrafficRanking)
				adminDashboard.GET("/invite-ranking", dashboardHandler.GetInviteRanking)
				adminDashboard.GET("/commission-stats", dashboardHandler.GetCommissionStats)
				adminDashboard.GET("/comprehensive-stats", dashboardHandler.GetComprehensiveStats)
			}

			// 系统设置
			adminSettings := admin.Group("/settings")
			{
				adminSettings.GET("", settingHandler.GetSettings)
				adminSettings.GET("/group/:group", settingHandler.GetSettingsByGroup)
				adminSettings.GET("/mappings", settingHandler.GetConfigMappings)
				adminSettings.PUT("", settingHandler.UpdateSettings)
				adminSettings.PUT("/:key", settingHandler.UpdateSetting)
				adminSettings.DELETE("/:key", settingHandler.DeleteSetting)
				adminSettings.POST("/test-email", settingHandler.TestSendEmail)
			}

			// 公告管理
			adminNotices := admin.Group("/notices")
			{
				adminNotices.GET("", noticeHandler.ListNotices)
				adminNotices.POST("", noticeHandler.CreateNotice)
				adminNotices.GET("/:id", noticeHandler.GetNotice)
				adminNotices.PUT("/:id", noticeHandler.UpdateNotice)
				adminNotices.DELETE("/:id", noticeHandler.DeleteNotice)
			}

			// 知识库管理
			adminKnowledges := admin.Group("/knowledges")
			{
				adminKnowledges.GET("", knowledgeHandler.ListKnowledges)
				adminKnowledges.POST("", knowledgeHandler.CreateKnowledge)
				adminKnowledges.GET("/:id", knowledgeHandler.GetKnowledge)
				adminKnowledges.PUT("/:id", knowledgeHandler.UpdateKnowledge)
				adminKnowledges.DELETE("/:id", knowledgeHandler.DeleteKnowledge)
			}

			// 邮件模板管理
			adminMailTemplates := admin.Group("/mail-templates")
			{
				adminMailTemplates.GET("", mailTemplateHandler.ListMailTemplates)
				adminMailTemplates.POST("", mailTemplateHandler.CreateMailTemplate)
				adminMailTemplates.GET("/:id", mailTemplateHandler.GetMailTemplate)
				adminMailTemplates.PUT("/:id", mailTemplateHandler.UpdateMailTemplate)
				adminMailTemplates.DELETE("/:id", mailTemplateHandler.DeleteMailTemplate)
			}

			// 优惠券管理
			adminCoupons := admin.Group("/coupons")
			{
				adminCoupons.GET("", couponHandler.ListCoupons)
				adminCoupons.POST("", couponHandler.CreateCoupon)
				adminCoupons.GET("/:id", couponHandler.GetCoupon)
				adminCoupons.PUT("/:id", couponHandler.UpdateCoupon)
				adminCoupons.DELETE("/:id", couponHandler.DeleteCoupon)
			}

			// 礼品卡模板管理
			adminGiftCardTemplates := admin.Group("/gift-card-templates")
			{
				adminGiftCardTemplates.GET("", giftCardHandler.ListTemplates)
				adminGiftCardTemplates.POST("", giftCardHandler.CreateTemplate)
				adminGiftCardTemplates.GET("/:id", giftCardHandler.GetTemplate)
				adminGiftCardTemplates.PUT("/:id", giftCardHandler.UpdateTemplate)
				adminGiftCardTemplates.DELETE("/:id", giftCardHandler.DeleteTemplate)
			}

			// 礼品码管理
			adminGiftCardCodes := admin.Group("/gift-card-codes")
			{
				adminGiftCardCodes.GET("", giftCardHandler.ListCodes)
				adminGiftCardCodes.POST("/generate", giftCardHandler.GenerateCodes)
				adminGiftCardCodes.GET("/:id", giftCardHandler.GetCode)
				adminGiftCardCodes.DELETE("/:id", giftCardHandler.DeleteCode)
			}

			// 邀请码管理
			adminInviteCodes := admin.Group("/invite-codes")
			{
				adminInviteCodes.GET("", inviteHandler.ListInviteCodes)
				adminInviteCodes.POST("", inviteHandler.CreateInviteCode)
				adminInviteCodes.GET("/:id", inviteHandler.GetInviteCode)
				adminInviteCodes.PUT("/:id", inviteHandler.UpdateInviteCode)
				adminInviteCodes.DELETE("/:id", inviteHandler.DeleteInviteCode)
			}

			// 佣金管理
			adminCommissions := admin.Group("/commissions")
			{
				adminCommissions.GET("", inviteHandler.ListCommissionLogsAdmin)
				adminCommissions.POST("/:id/settle", inviteHandler.SettleCommission)
			}

			// 服务器分组管理
			adminServerGroups := admin.Group("/server-groups")
			{
				adminServerGroups.GET("", serverGroupHandler.ListServerGroups)
				adminServerGroups.POST("", serverGroupHandler.CreateServerGroup)
				adminServerGroups.GET("/all", serverGroupHandler.ListAllServerGroups)
				adminServerGroups.GET("/:id", serverGroupHandler.GetServerGroup)
				adminServerGroups.PUT("/:id", serverGroupHandler.UpdateServerGroup)
				adminServerGroups.DELETE("/:id", serverGroupHandler.DeleteServerGroup)
			}

			// 服务器路由管理
			adminServerRoutes := admin.Group("/server-routes")
			{
				adminServerRoutes.GET("", serverRouteHandler.ListServerRoutes)
				adminServerRoutes.POST("", serverRouteHandler.CreateServerRoute)
				adminServerRoutes.GET("/:id", serverRouteHandler.GetServerRoute)
				adminServerRoutes.PUT("/:id", serverRouteHandler.UpdateServerRoute)
				adminServerRoutes.DELETE("/:id", serverRouteHandler.DeleteServerRoute)
				adminServerRoutes.GET("/group/:group_id", serverRouteHandler.ListServerRoutesByGroup)
			}

			// 服务器机器管理
			adminServerMachines := admin.Group("/server-machines")
			{
				adminServerMachines.GET("", serverMachineHandler.ListServerMachines)
				adminServerMachines.POST("", serverMachineHandler.CreateServerMachine)
				adminServerMachines.GET("/all", serverMachineHandler.ListAllServerMachines)
				adminServerMachines.GET("/:id", serverMachineHandler.GetServerMachine)
				adminServerMachines.PUT("/:id", serverMachineHandler.UpdateServerMachine)
				adminServerMachines.DELETE("/:id", serverMachineHandler.DeleteServerMachine)
				adminServerMachines.PUT("/:id/status", serverMachineHandler.UpdateServerMachineStatus)
				adminServerMachines.PUT("/:id/load", serverMachineHandler.UpdateServerMachineLoad)
				adminServerMachines.POST("/:id/reset-token", serverMachineHandler.ResetToken)
				adminServerMachines.GET("/:id/install-command", serverMachineHandler.GetInstallCommand)
			}

			// 插件管理
			adminPlugins := admin.Group("/plugins")
			{
				adminPlugins.GET("", pluginHandler.ListPlugins)
				adminPlugins.POST("", pluginHandler.CreatePlugin)
				adminPlugins.GET("/:id", pluginHandler.GetPlugin)
				adminPlugins.PUT("/:id", pluginHandler.UpdatePlugin)
				adminPlugins.DELETE("/:id", pluginHandler.DeletePlugin)
				adminPlugins.PUT("/:id/status", pluginHandler.UpdatePluginStatus)
				adminPlugins.POST("/:id/enable", pluginHandler.EnablePlugin)
				adminPlugins.POST("/:id/disable", pluginHandler.DisablePlugin)
			}

			// 审计日志管理
			adminAuditLogs := admin.Group("/audit-logs")
			{
				adminAuditLogs.GET("", auditLogHandler.ListAuditLogs)
				adminAuditLogs.GET("/:id", auditLogHandler.GetAuditLog)
				adminAuditLogs.DELETE("/:id", auditLogHandler.DeleteAuditLog)
			}

			// 支付网关管理
			adminPaymentGateways := admin.Group("/payment/gateways")
			{
				adminPaymentGateways.GET("", paymentGatewayHandler.ListGateways)
				adminPaymentGateways.POST("", paymentGatewayHandler.CreateGateway)
				adminPaymentGateways.GET("/:id", paymentGatewayHandler.GetGateway)
				adminPaymentGateways.PUT("/:id", paymentGatewayHandler.UpdateGateway)
				adminPaymentGateways.DELETE("/:id", paymentGatewayHandler.DeleteGateway)
				adminPaymentGateways.PUT("/:id/status", paymentGatewayHandler.UpdateGatewayStatus)
				adminPaymentGateways.PUT("/:id/sort", paymentGatewayHandler.UpdateGatewaySort)
			}

			// Telegram 管理
			adminTelegram := admin.Group("/telegram")
			{
				adminTelegram.POST("/set-webhook", telegramHandler.SetWebhook)
			}

			// 系统管理
			adminSystem := admin.Group("/system")
			{
				adminSystem.GET("/status", systemHandler.GetSystemStatus)
				adminSystem.GET("/info", systemHandler.GetSystemInfo)
				adminSystem.GET("/scheduler", systemHandler.GetSchedulerStatus)
				adminSystem.GET("/check-update", updateHandler.CheckUpdate)
				adminSystem.POST("/execute-update", updateHandler.ExecuteUpdate)
			}
		}

		// 公开设置接口（无需认证）
		publicSettings := apiV1.Group("/settings")
		{
			publicSettings.GET("/public", settingHandler.GetPublicSettings)
		}

		// 公共配置接口
		apiV1.GET("/config", commHandler.GetConfig)

		// 公开公告接口（无需认证）
		publicNotices := apiV1.Group("/notices")
		{
			publicNotices.GET("", noticeHandler.ListVisibleNotices)
		}

		// 公开支付方式列表（无需认证）
		publicPaymentGateways := apiV1.Group("/payment/gateways")
		{
			publicPaymentGateways.GET("", paymentGatewayHandler.ListGateways)
		}

		// 公开知识库接口（无需认证）
		publicKnowledges := apiV1.Group("/knowledges")
		{
			publicKnowledges.GET("", knowledgeHandler.ListVisibleKnowledges)
			publicKnowledges.GET("/categories", knowledgeHandler.GetCategories)
		}

		// 用户优惠券接口（需要认证）
		userCoupons := apiV1.Group("/coupons")
		userCoupons.Use(middleware.JWTAuth())
		{
			userCoupons.POST("/validate", couponHandler.ValidateCoupon)
		}

		// 用户礼品卡接口（需要认证）
		userGiftCards := apiV1.Group("/gift-cards")
		userGiftCards.Use(middleware.JWTAuth())
		{
			userGiftCards.POST("/use", giftCardHandler.UseCode)
		}

		// 用户邀请码接口（需要认证）
		userInvite := apiV1.Group("/invite")
		userInvite.Use(middleware.JWTAuth())
		{
			userInvite.GET("/code", inviteHandler.GetUserInviteCode)
			userInvite.POST("/use", inviteHandler.UseInviteCode)
			userInvite.GET("/commission/stats", inviteHandler.GetCommissionStats)
			userInvite.GET("/commission/logs", inviteHandler.ListCommissionLogs)
			userInvite.POST("/commission/withdraw", inviteHandler.WithdrawCommission)
		}
	}

	// 节点 REST API 路由（兼容 Xboard-Node）
	// 使用 NodeAPIKeyAuth 中间件认证，与 gRPC 节点认证方式一致
	nodeAPI := r.Group("")
	nodeAPI.Use(middleware.NodeAPIKeyAuth())
	{
		// v2 握手
		nodeAPI.POST("/api/v2/server/handshake", nodeServerHandler.Handshake)
		// v2 综合上报
		nodeAPI.POST("/api/v2/server/report", nodeServerHandler.Report)

		// v1 legacy 端点（兼容旧版节点）
		nodeAPI.GET("/api/v1/server/UniProxy/config", nodeServerHandler.GetConfig)
		nodeAPI.GET("/api/v1/server/UniProxy/user", nodeServerHandler.GetUsers)
		nodeAPI.POST("/api/v1/server/UniProxy/push", nodeServerHandler.PushTraffic)
		nodeAPI.POST("/api/v1/server/UniProxy/alive", nodeServerHandler.PushAlive)
		nodeAPI.POST("/api/v1/server/UniProxy/status", nodeServerHandler.PushStatus)

		// v2 config/user（与 v1 相同逻辑）
		nodeAPI.GET("/api/v2/server/config", nodeServerHandler.GetConfig)
		nodeAPI.GET("/api/v2/server/user", nodeServerHandler.GetUsers)

		// machine mode 端点
		nodeAPI.POST("/api/v2/server/machine/nodes", nodeServerHandler.GetMachineNodes)
		nodeAPI.POST("/api/v2/server/machine/status", nodeServerHandler.ReportMachineStatus)
	}

	// 节点 WebSocket 端点（Xboard-Node 实时推送）
	// 认证在 handler 内部处理（WS 不走中间件）
	r.GET("/api/v2/server/ws", nodeWSHandler.Handle)

	// 前端 SPA：未匹配 API 路由的请求交给 SPA 处理
	// 静态资源直接从 embed.FS 提供，其余路径回退到 index.html
	r.NoRoute(static.SPAHandler())

	return r
}
