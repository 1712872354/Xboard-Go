package ratelimit

import (
	"sync"
	"time"
)

// memoryLimiter 内存版限流器（Redis 不可用时的降级方案）
// 使用令牌桶算法，单节点有效，不支持分布式
type memoryLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	prefix  string
}

// tokenBucket 令牌桶
type tokenBucket struct {
	tokens     float64   // 当前令牌数
	capacity   float64   // 桶容量（最大令牌数）
	rate       float64   // 每秒生成速率
	lastRefill time.Time // 上次补充时间
}

// NewMemoryLimiter 创建内存限流器
func NewMemoryLimiter(prefix string) Limiter {
	if prefix == "" {
		prefix = "rate_limit"
	}
	return &memoryLimiter{
		buckets: make(map[string]*tokenBucket),
		prefix:  prefix,
	}
}

// Allow 检查是否允许请求通过
func (l *memoryLimiter) Allow(key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	if limit <= 0 || window <= 0 {
		return true, 0, time.Time{}, nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	bucketKey := l.prefix + ":" + key
	bucket, exists := l.buckets[bucketKey]

	// 计算令牌生成速率（每秒生成多少令牌）
	ratePerSecond := float64(limit) / window.Seconds()

	if !exists {
		// 新桶，初始满令牌
		bucket = &tokenBucket{
			tokens:     float64(limit),
			capacity:   float64(limit),
			rate:       ratePerSecond,
			lastRefill: time.Now(),
		}
		l.buckets[bucketKey] = bucket
	}

	// 补充令牌
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	if elapsed > 0 {
		bucket.tokens += elapsed * bucket.rate
		if bucket.tokens > bucket.capacity {
			bucket.tokens = bucket.capacity
		}
		bucket.lastRefill = now
	}

	// 检查是否有足够令牌
	if bucket.tokens >= 1 {
		bucket.tokens -= 1

		remaining := int(bucket.tokens)
		// 计算重置时间（需要多久才能恢复到满）
		missingTokens := bucket.capacity - bucket.tokens
		resetDuration := time.Duration(missingTokens/bucket.rate) * time.Second
		resetTime := now.Add(resetDuration)

		return true, remaining, resetTime, nil
	}

	// 令牌不足，计算下次可用时间
	needed := 1.0 - bucket.tokens
	waitDuration := time.Duration(needed/bucket.rate) * time.Second
	resetTime := now.Add(waitDuration)

	return false, 0, resetTime, nil
}
