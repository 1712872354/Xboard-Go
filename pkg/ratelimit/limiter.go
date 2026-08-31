package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter 限流器接口
type Limiter interface {
	// Allow 检查是否允许请求通过
	// key: 限流键（如 IP、用户ID等）
	// limit: 时间窗口内最大请求数
	// window: 时间窗口大小
	// 返回: 是否允许, 剩余请求数, 重置时间, 错误
	Allow(key string, limit int, window time.Duration) (bool, int, time.Time, error)
}

// redisLimiter 基于 Redis 滑动窗口的限流器
// 使用 Redis ZSET 实现滑动窗口算法，保证原子性和分布式一致性
type redisLimiter struct {
	client *redis.Client
	prefix string
}

// NewRedisLimiter 创建 Redis 限流器
func NewRedisLimiter(client *redis.Client, prefix string) Limiter {
	if prefix == "" {
		prefix = "rate_limit"
	}
	return &redisLimiter{
		client: client,
		prefix: prefix,
	}
}

// slidingWindowLua 滑动窗口限流 Lua 脚本
// 原子操作：移除过期记录 → 检查当前窗口请求数 → 添加当前请求
// KEYS[1]: ZSET key
// ARGV[1]: 当前时间戳（毫秒）
// ARGV[2]: 窗口大小（毫秒）
// ARGV[3]: 最大请求数
// ARGV[4]: 请求唯一标识（时间戳+随机数）
// 返回: 0=限流, 1=通过
var slidingWindowLua = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local requestId = ARGV[4]

-- 计算窗口起始时间
local windowStart = now - window

-- 移除窗口外的过期记录
redis.call('ZREMRANGEBYSCORE', key, 0, windowStart)

-- 获取当前窗口内的请求数
local count = redis.call('ZCARD', key)

-- 如果超过限制，返回 0（被限流）
if count >= limit then
    return 0
end

-- 添加当前请求
redis.call('ZADD', key, now, requestId)

-- 设置过期时间（窗口大小 + 缓冲时间）
redis.call('EXPIRE', key, math.ceil(window / 1000) + 1)

-- 返回 1（允许通过）
return 1
`)

// Allow 检查是否允许请求通过
func (l *redisLimiter) Allow(key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	if limit <= 0 || window <= 0 {
		return true, 0, time.Time{}, nil
	}

	ctx := context.Background()
	now := time.Now()
	nowMs := now.UnixMilli()
	windowMs := window.Milliseconds()

	// 生成唯一请求 ID（时间戳 + 纳秒，保证同一毫秒内也唯一）
	requestID := fmt.Sprintf("%d-%d", nowMs, now.Nanosecond())

	redisKey := fmt.Sprintf("%s:%s", l.prefix, key)

	// 执行 Lua 脚本
	result, err := slidingWindowLua.Run(ctx, l.client, []string{redisKey},
		nowMs, windowMs, limit, requestID).Int()

	if err != nil {
		// Redis 出错时默认放行（避免 Redis 故障导致服务不可用）
		return true, limit, now.Add(window), nil
	}

	allowed := result == 1

	// 计算剩余请求数和重置时间
	// 注意：为了性能不每次都查剩余数，可以用另一个脚本或估算
	// 这里简化处理
	remaining := 0
	resetTime := now.Add(window)

	if allowed {
		// 如果允许，剩余 = limit - 当前数 - 1（刚加的）
		// 简化：返回一个估算值
		remaining = limit - 1
	} else {
		remaining = 0
		// 计算最早的请求何时过期（窗口重置时间）
		// 简化处理：返回当前时间 + 窗口的 1/10 作为下一次重试时间
	}

	return allowed, remaining, resetTime, nil
}

// GetRemaining 获取剩余请求数（用于返回给客户端）
func (l *redisLimiter) GetRemaining(key string, limit int, window time.Duration) (int, time.Time, error) {
	ctx := context.Background()
	now := time.Now()
	nowMs := now.UnixMilli()
	windowMs := window.Milliseconds()
	windowStart := nowMs - windowMs

	redisKey := fmt.Sprintf("%s:%s", l.prefix, key)

	// 移除过期记录
	l.client.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart))

	// 获取当前请求数
	count, err := l.client.ZCard(ctx, redisKey).Result()
	if err != nil {
		return limit, now.Add(window), nil
	}

	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return remaining, now.Add(window), nil
}
