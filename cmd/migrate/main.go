package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"xboard-go/config"
	"xboard-go/pkg/database"
)

func main() {
	configPath := flag.String("c", "config.yaml", "配置文件路径")
	action := flag.String("action", "up", "迁移操作: up, down, status")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	err = database.Init(&cfg.Database)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	db := database.Get()

	switch *action {
	case "up":
		fmt.Println("执行数据库迁移...")
		if err := db.AutoMigrate(
			// 在这里添加需要迁移的模型
			// &model.User{},
			// &model.Plan{},
			// &model.Order{},
		); err != nil {
			log.Fatalf("迁移失败: %v", err)
		}
		fmt.Println("迁移完成")

	case "down":
		fmt.Println("回滚数据库迁移...")
		fmt.Println("注意: AutoMigrate 不支持自动回滚，请手动处理")

	case "status":
		fmt.Println("数据库迁移状态:")
		// 检查表是否存在
		tables := []string{
			"users", "plans", "orders", "nodes", "tickets",
			"settings", "notices", "knowledges", "coupons",
			"gift_card_templates", "gift_card_codes", "invite_codes",
			"commission_logs", "redeem_codes", "admin_audit_logs", "plugins",
		}

		for _, table := range tables {
			if db.Migrator().HasTable(table) {
				fmt.Printf("  ✓ %s\n", table)
			} else {
				fmt.Printf("  ✗ %s (不存在)\n", table)
			}
		}

	default:
		fmt.Printf("未知操作: %s\n", *action)
		os.Exit(1)
	}
}
