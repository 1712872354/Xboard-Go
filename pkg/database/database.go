package database

import (
	"fmt"
	"time"

	"xboard-go/config"
	applogger "xboard-go/pkg/logger"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var db *gorm.DB

// Init 初始化数据库连接
func Init(cfg *config.DatabaseConfig) error {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "postgres":
		dialector = postgres.Open(cfg.DSN())
	case "mysql":
		dialector = mysql.Open(cfg.DSN())
	case "sqlite", "sqlite3":
		dialector = sqlite.Open(cfg.DSN())
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	}

	var err error
	db, err = gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	applogger.Sugar().Infof("Database connected: %s@%s:%d/%s", cfg.User, cfg.Host, cfg.Port, cfg.DBName)
	return nil
}

// Get 获取数据库连接
func Get() *gorm.DB {
	if db == nil {
		panic("database not initialized")
	}
	return db
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate(models ...interface{}) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.AutoMigrate(models...)
}
