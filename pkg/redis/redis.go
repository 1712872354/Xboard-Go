package redis

import (
	"context"
	"fmt"
	"time"

	"xboard-go/config"
	"xboard-go/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

// Init 初始化 Redis 连接
func Init(cfg *config.RedisConfig) error {
	client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	logger.Sugar().Infof("Redis connected: %s (db=%d)", cfg.Addr(), cfg.DB)
	return nil
}

// Client 获取 Redis 客户端
func Client() *redis.Client {
	if client == nil {
		panic("redis not initialized")
	}
	return client
}

// Set 设置键值
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func Get(ctx context.Context, key string) (string, error) {
	return client.Get(ctx, key).Result()
}

// Del 删除键
func Del(ctx context.Context, keys ...string) error {
	return client.Del(ctx, keys...).Err()
}

// Exists 判断键是否存在
func Exists(ctx context.Context, key string) (bool, error) {
	result, err := client.Exists(ctx, key).Result()
	return result > 0, err
}
