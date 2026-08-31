package service

import (
	"context"
	"fmt"
	"time"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/logger"
	appredis "xboard-go/pkg/redis"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// TrafficService 流量服务接口
type TrafficService interface {
	// AddTraffic 增加用户已用流量（Redis 原子操作）
	AddTraffic(ctx context.Context, userID uint, upload, download int64, nodeID uint) (bool, error)
	// GetUsedTraffic 获取用户已用流量（优先从 Redis 获取）
	GetUsedTraffic(ctx context.Context, userID uint) (int64, error)
	// SyncToDB 将 Redis 中的流量数据同步到数据库
	SyncToDB(ctx context.Context, userID uint) error
	// SyncAllToDB 同步所有用户的流量数据到数据库
	SyncAllToDB(ctx context.Context) (int, error)
	// GetUserTrafficStats 获取用户流量统计（今日/本周/本月）
	GetUserTrafficStats(ctx context.Context, userID uint) (*TrafficStats, error)
	// GetUserTrafficHistory 获取用户流量历史记录
	GetUserTrafficHistory(ctx context.Context, userID uint, page, pageSize int, start, end time.Time) ([]model.TrafficLog, int64, error)
	// GetDailyTraffic 获取按日聚合的流量数据（用于图表）
	GetDailyTraffic(ctx context.Context, userID uint, days int) ([]DailyTraffic, error)
	// GetDailyTrafficByDateRange 按日期范围获取每日流量统计
	GetDailyTrafficByDateRange(ctx context.Context, userID uint, start, end time.Time) ([]DailyTraffic, error)
}

type trafficService struct {
	db *gorm.DB
}

// NewTrafficService 创建流量服务
func NewTrafficService() TrafficService {
	return &trafficService{
		db: database.Get(),
	}
}

// Redis Key 前缀
const (
	trafficKeyPrefix  = "traffic:user:"    // 流量累计值 Key
	trafficSyncSetKey = "traffic:sync:set" // 需要同步的用户集合
)

// Lua 脚本：原子增加流量并检查是否超限
// KEYS[1] = traffic key
// KEYS[2] = sync set key
// ARGV[1] = 增加的流量值
// ARGV[2] = 流量限制（0 表示不限）
// ARGV[3] = 用户ID
// 返回：1=成功（未超限）, 0=超出限制
var addTrafficLua = redis.NewScript(`
local key = KEYS[1]
local syncSet = KEYS[2]
local delta = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local userID = ARGV[3]

-- 增加已用流量
local current = redis.call('INCRBY', key, delta)

-- 加入待同步集合
redis.call('SADD', syncSet, userID)

-- 设置过期时间（24小时，防止残留）
redis.call('EXPIRE', key, 86400)

-- 检查是否超出限制
if limit > 0 and current >= limit then
    return 0  -- 超出限制
end

return 1  -- 成功（未超限）
`)

// AddTraffic 增加用户已用流量
// 返回值：isOverLimit（是否超出限制）, error
func (s *trafficService) AddTraffic(ctx context.Context, userID uint, upload, download int64, nodeID uint) (bool, error) {
	totalDelta := upload + download
	if totalDelta <= 0 {
		return false, nil
	}

	trafficKey := fmt.Sprintf("%s%d", trafficKeyPrefix, userID)

	// 获取流量限制（带缓存）
	limit, err := s.getTrafficLimit(ctx, userID)
	if err != nil {
		logger.Sugar().Warnf("Failed to get traffic limit for user %d: %v", userID, err)
		limit = 0
	}

	// 执行 Lua 脚本
	result, err := addTrafficLua.Run(ctx, appredis.Client(),
		[]string{trafficKey, trafficSyncSetKey},
		totalDelta, limit, userID,
	).Int()

	if err != nil {
		// Redis 失败，降级到直接写数据库
		logger.Sugar().Warnf("Redis add traffic failed, fallback to DB: user_id=%d, err=%v", userID, err)
		if err := s.addTrafficToDB(userID, upload, download); err != nil {
			return false, fmt.Errorf("DB fallback also failed: %w", err)
		}
		// 检查是否超流量
		user, _ := s.getUserFromDB(userID)
		if user != nil && user.TrafficLimit > 0 && user.UsedTraffic >= user.TrafficLimit {
			return true, nil
		}
		return false, nil
	}

	isOverLimit := result == 0

	// 如果超出限制，禁用用户
	if isOverLimit {
		go s.disableUser(userID)
	}

	return isOverLimit, nil
}

// GetUsedTraffic 获取用户已用流量
func (s *trafficService) GetUsedTraffic(ctx context.Context, userID uint) (int64, error) {
	trafficKey := fmt.Sprintf("%s%d", trafficKeyPrefix, userID)

	// 从 Redis 获取增量
	val, err := appredis.Get(ctx, trafficKey)
	var redisDelta int64
	if err == nil {
		fmt.Sscanf(val, "%d", &redisDelta)
	}

	// 从数据库获取基准值
	user, err := s.getUserFromDB(userID)
	if err != nil {
		return 0, err
	}

	// 总已用流量 = 数据库基准 + Redis 增量
	return user.UsedTraffic + redisDelta, nil
}

// SyncToDB 将单个用户的 Redis 流量数据同步到数据库
func (s *trafficService) SyncToDB(ctx context.Context, userID uint) error {
	trafficKey := fmt.Sprintf("%s%d", trafficKeyPrefix, userID)

	// 原子获取并删除（使用 GETDEL 或先 GET 再 DEL）
	val, err := appredis.Client().GetDel(ctx, trafficKey).Result()
	if err == redis.Nil {
		return nil // 没有数据，无需同步
	}
	if err != nil {
		return fmt.Errorf("failed to get traffic from redis: %w", err)
	}

	var delta int64
	fmt.Sscanf(val, "%d", &delta)
	if delta <= 0 {
		// 移除同步集合
		appredis.Client().SRem(ctx, trafficSyncSetKey, fmt.Sprintf("%d", userID))
		return nil
	}

	// 同步到数据库（使用事务）
	err = s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.User{}).
			Where("id = ?", userID).
			UpdateColumn("used_traffic",
				gorm.Expr("used_traffic + ?", delta))
		if result.Error != nil {
			return result.Error
		}

		// 记录流量日志
		log := model.TrafficLog{
			UserID:     userID,
			NodeID:     0, // 批量同步，不区分节点
			Upload:     0,
			Download:   delta,
			RecordedAt: time.Now(),
		}
		return tx.Create(&log).Error
	})

	if err != nil {
		// 同步失败，把数据写回 Redis
		appredis.Set(ctx, trafficKey, fmt.Sprintf("%d", delta), 24*time.Hour)
		return fmt.Errorf("failed to sync traffic to DB: %w", err)
	}

	// 同步成功，从同步集合中移除
	appredis.Client().SRem(ctx, trafficSyncSetKey, fmt.Sprintf("%d", userID))

	logger.Sugar().Debugf("Synced traffic for user %d: %d bytes", userID, delta)
	return nil
}

// SyncAllToDB 同步所有待同步用户的流量数据
func (s *trafficService) SyncAllToDB(ctx context.Context) (int, error) {
	// 获取所有需要同步的用户ID
	members, err := appredis.Client().SMembers(ctx, trafficSyncSetKey).Result()
	if err != nil {
		return 0, err
	}

	if len(members) == 0 {
		return 0, nil
	}

	synced := 0
	for _, member := range members {
		var userID uint
		fmt.Sscanf(member, "%d", &userID)
		if userID == 0 {
			continue
		}

		if err := s.SyncToDB(ctx, userID); err != nil {
			logger.Sugar().Errorf("Failed to sync traffic for user %d: %v", userID, err)
			continue
		}
		synced++
	}

	logger.Sugar().Infof("Traffic sync completed: %d/%d users synced", synced, len(members))
	return synced, nil
}

// TrafficStats 流量统计
type TrafficStats struct {
	TodayUpload    int64 `json:"today_upload"`
	TodayDownload  int64 `json:"today_download"`
	WeekUpload     int64 `json:"week_upload"`
	WeekDownload   int64 `json:"week_download"`
	MonthUpload    int64 `json:"month_upload"`
	MonthDownload  int64 `json:"month_download"`
	TotalUsed      int64 `json:"total_used"`
	TrafficLimit   int64 `json:"traffic_limit"`
	Remaining      int64 `json:"remaining"`
}

// GetUserTrafficStats 获取用户流量统计
func (s *trafficService) GetUserTrafficStats(ctx context.Context, userID uint) (*TrafficStats, error) {
	user, err := s.getUserFromDB(userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// 今日
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayUpload, todayDownload, err := s.sumTraffic(userID, todayStart, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get today traffic: %w", err)
	}

	// 本周（周一到今天）
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	weekUpload, weekDownload, err := s.sumTraffic(userID, weekStart, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get week traffic: %w", err)
	}

	// 本月
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthUpload, monthDownload, err := s.sumTraffic(userID, monthStart, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get month traffic: %w", err)
	}

	// 总已用流量（数据库 + Redis 增量）
	totalUsed, err := s.GetUsedTraffic(ctx, userID)
	if err != nil {
		totalUsed = user.UsedTraffic
	}

	var remaining int64
	if user.TrafficLimit > 0 {
		remaining = user.TrafficLimit - totalUsed
		if remaining < 0 {
			remaining = 0
		}
	} else {
		remaining = -1 // -1 表示不限流量
	}

	return &TrafficStats{
		TodayUpload:   todayUpload,
		TodayDownload: todayDownload,
		WeekUpload:    weekUpload,
		WeekDownload:  weekDownload,
		MonthUpload:   monthUpload,
		MonthDownload: monthDownload,
		TotalUsed:     totalUsed,
		TrafficLimit:  user.TrafficLimit,
		Remaining:     remaining,
	}, nil
}

// GetUserTrafficHistory 获取用户流量历史记录
func (s *trafficService) GetUserTrafficHistory(ctx context.Context, userID uint, page, pageSize int, start, end time.Time) ([]model.TrafficLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var logs []model.TrafficLog
	var total int64

	query := s.db.Model(&model.TrafficLog{}).Where("user_id = ?", userID)

	if !start.IsZero() {
		query = query.Where("recorded_at >= ?", start)
	}
	if !end.IsZero() {
		query = query.Where("recorded_at <= ?", end)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("recorded_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// sumTraffic 统计用户在时间段内的流量总和
func (s *trafficService) sumTraffic(userID uint, start, end time.Time) (int64, int64, error) {
	var result struct {
		Upload   int64
		Download int64
	}

	err := s.db.Model(&model.TrafficLog{}).
		Select("COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download").
		Where("user_id = ? AND recorded_at >= ? AND recorded_at <= ?", userID, start, end).
		Scan(&result).Error

	return result.Upload, result.Download, err
}

// getTrafficLimit 获取用户流量限制（带缓存）
func (s *trafficService) getTrafficLimit(ctx context.Context, userID uint) (int64, error) {
	limitKey := fmt.Sprintf("%s%d:limit", trafficKeyPrefix, userID)

	val, err := appredis.Get(ctx, limitKey)
	if err == nil {
		var limit int64
		fmt.Sscanf(val, "%d", &limit)
		return limit, nil
	}

	// 从数据库获取
	user, err := s.getUserFromDB(userID)
	if err != nil {
		return 0, err
	}

	// 缓存 1 小时
	_ = appredis.Set(ctx, limitKey, fmt.Sprintf("%d", user.TrafficLimit), time.Hour)

	return user.TrafficLimit, nil
}

// getUserFromDB 从数据库获取用户
func (s *trafficService) getUserFromDB(userID uint) (*model.User, error) {
	var user model.User
	err := s.db.First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// addTrafficToDB 直接写数据库（降级方案）
func (s *trafficService) addTrafficToDB(userID uint, upload, download int64) error {
	total := upload + download
	if total <= 0 {
		return nil
	}

	return s.db.Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("used_traffic",
			gorm.Expr("used_traffic + ?", total)).Error
}

// disableUser 禁用用户
func (s *trafficService) disableUser(userID uint) {
	result := s.db.Model(&model.User{}).
		Where("id = ? AND status = ?", userID, 1).
		Update("status", 0)
	if result.Error != nil {
		logger.Sugar().Errorf("Failed to disable user %d: %v", userID, result.Error)
		return
	}
	if result.RowsAffected > 0 {
		logger.Sugar().Infof("User %d disabled due to traffic over limit", userID)
	}
}

// DailyTraffic 按日聚合的流量统计
type DailyTraffic struct {
	Date     string `json:"date"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

// GetDailyTraffic 获取按日聚合的流量数据
func (s *trafficService) GetDailyTraffic(ctx context.Context, userID uint, days int) ([]DailyTraffic, error) {
	if days <= 0 || days > 90 {
		days = 7
	}

	startDate := time.Now().AddDate(0, 0, -days+1).Truncate(24 * time.Hour)

	type result struct {
		Date     string
		Upload   int64
		Download int64
	}

	var results []result
	if err := s.db.Model(&model.TrafficLog{}).
		Select("DATE(recorded_at) as date, COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download").
		Where("user_id = ? AND recorded_at >= ?", userID, startDate).
		Group("DATE(recorded_at)").
		Order("date ASC").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	// 补全缺失的日期
	resultMap := make(map[string]*result)
	for i := range results {
		resultMap[results[i].Date] = &results[i]
	}

	var daily []DailyTraffic
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")

		if r, exists := resultMap[dateStr]; exists {
			daily = append(daily, DailyTraffic{
				Date:     dateStr,
				Upload:   r.Upload,
				Download: r.Download,
			})
		} else {
			daily = append(daily, DailyTraffic{
				Date:     dateStr,
				Upload:   0,
				Download: 0,
			})
		}
	}

	return daily, nil
}

// GetDailyTrafficByDateRange 按日期范围获取每日流量统计
func (s *trafficService) GetDailyTrafficByDateRange(ctx context.Context, userID uint, start, end time.Time) ([]DailyTraffic, error) {
	if end.Before(start) {
		start, end = end, start
	}

	startDate := start.Truncate(24 * time.Hour)
	endDate := end.Add(24 * time.Hour).Truncate(24 * time.Hour)

	type result struct {
		Date     string
		Upload   int64
		Download int64
	}

	var results []result
	if err := s.db.Model(&model.TrafficLog{}).
		Select("DATE(recorded_at) as date, COALESCE(SUM(upload), 0) as upload, COALESCE(SUM(download), 0) as download").
		Where("user_id = ? AND recorded_at >= ? AND recorded_at < ?", userID, startDate, endDate).
		Group("DATE(recorded_at)").
		Order("date ASC").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	resultMap := make(map[string]*result)
	for i := range results {
		resultMap[results[i].Date] = &results[i]
	}

	var daily []DailyTraffic
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		if r, exists := resultMap[dateStr]; exists {
			daily = append(daily, DailyTraffic{Date: dateStr, Upload: r.Upload, Download: r.Download})
		} else {
			daily = append(daily, DailyTraffic{Date: dateStr, Upload: 0, Download: 0})
		}
	}

	return daily, nil
}
