package service

import (
	"encoding/json"
	"fmt"
	"time"

	"xboard-go/internal/model"
	"xboard-go/pkg/database"
	"xboard-go/pkg/logger"

	"gorm.io/gorm"
)

// TrafficResetService 流量重置服务接口
type TrafficResetService interface {
	// CheckAndReset 检查用户是否需要重置流量，需要则执行
	CheckAndReset(user *model.User, triggerSource string) bool
	// PerformReset 执行流量重置
	PerformReset(user *model.User, triggerSource string) bool
	// CalculateNextResetTime 计算下次重置时间
	CalculateNextResetTime(user *model.User, plan *model.Plan) *time.Time
	// BatchCheckReset 批量检查并重置用户流量
	BatchCheckReset(batchSize int) (int, int, error)
	// SetInitialResetTime 设置初始重置时间
	SetInitialResetTime(user *model.User, plan *model.Plan)
	// ManualReset 管理员手动重置
	ManualReset(user *model.User) bool
}

type trafficResetService struct {
	db *gorm.DB
}

// NewTrafficResetService 创建流量重置服务
func NewTrafficResetService() TrafficResetService {
	return &trafficResetService{
		db: database.Get(),
	}
}

// CheckAndReset 检查并重置流量
func (s *trafficResetService) CheckAndReset(user *model.User, triggerSource string) bool {
	if !user.ShouldResetTraffic() {
		return false
	}
	return s.PerformReset(user, triggerSource)
}

// PerformReset 执行流量重置
func (s *trafficResetService) PerformReset(user *model.User, triggerSource string) bool {
	oldUpload := int64(0)
	oldDownload := user.UsedTraffic // 简化：只记录总使用量
	oldTotal := user.UsedTraffic

	// 获取用户的套餐以计算下次重置时间
	var plan *model.Plan
	if user.PlanID != nil {
		plan = &model.Plan{}
		if err := s.db.First(plan, *user.PlanID).Error; err != nil {
			plan = nil
		}
	}

	nextResetTime := s.CalculateNextResetTime(user, plan)
	resetType := s.getResetTypeFromPlan(plan)

	now := time.Now()
	// 更新用户流量
	updates := map[string]interface{}{
		"used_traffic":  0,
		"last_reset_at": now,
		"reset_count":   gorm.Expr("reset_count + 1"),
	}
	if nextResetTime != nil {
		updates["next_reset_at"] = *nextResetTime
	} else {
		updates["next_reset_at"] = nil
	}

	if err := s.db.Model(user).Updates(updates).Error; err != nil {
		logger.Sugar().Errorf("Failed to reset traffic for user %d: %v", user.ID, err)
		return false
	}

	// 记录重置日志
	log := model.TrafficResetLog{
		UserID:        user.ID,
		ResetType:     resetType,
		TriggerSource: triggerSource,
		OldUpload:     oldUpload,
		OldDownload:   oldDownload,
		OldTotal:      oldTotal,
		NewUpload:     0,
		NewDownload:   0,
		NewTotal:      0,
		ResetTime:     now,
	}
	if err := s.db.Create(&log).Error; err != nil {
		logger.Sugar().Warnf("Failed to create traffic reset log for user %d: %v", user.ID, err)
	}

	logger.Sugar().Infof("Traffic reset for user %d (type=%s, source=%s, old_total=%d)",
		user.ID, resetType, triggerSource, oldTotal)
	return true
}

// CalculateNextResetTime 计算下次重置时间
func (s *trafficResetService) CalculateNextResetTime(user *model.User, plan *model.Plan) *time.Time {
	if plan == nil {
		return nil
	}

	resetMethod := plan.ResetTrafficMethod

	// 跟随系统设置（从 settings 表读取）
	if resetMethod == model.ResetTrafficFollowSystem {
		resetMethod = s.getSystemResetMethod()
	}

	if resetMethod == model.ResetTrafficNever || user.ExpiredAt == nil {
		return nil
	}

	now := time.Now()
	var next time.Time

	switch resetMethod {
	case model.ResetTrafficFirstDayMonth:
		next = s.getNextMonthFirstDay(now)
	case model.ResetTrafficMonthly:
		next = s.getNextMonthlyReset(user, now)
	case model.ResetTrafficFirstDayYear:
		next = s.getNextYearFirstDay(now)
	case model.ResetTrafficYearly:
		next = s.getNextYearlyReset(user, now)
	default:
		return nil
	}

	return &next
}

// getNextMonthFirstDay 获取下月1号
func (s *trafficResetService) getNextMonthFirstDay(from time.Time) time.Time {
	year, month, _ := from.Date()
	nextMonth := month + 1
	nextYear := year
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}
	return time.Date(nextYear, nextMonth, 1, 0, 0, 0, 0, from.Location())
}

// getNextMonthlyReset 按月重置（按开通日）
func (s *trafficResetService) getNextMonthlyReset(user *model.User, from time.Time) time.Time {
	if user.ExpiredAt == nil {
		return s.getNextMonthFirstDay(from)
	}

	expiredAt := *user.ExpiredAt
	resetDay := expiredAt.Day()
	resetHour := expiredAt.Hour()
	resetMin := expiredAt.Minute()
	resetSec := expiredAt.Second()

	// 尝试本月
	target := time.Date(from.Year(), from.Month(), resetDay, resetHour, resetMin, resetSec, 0, from.Location())
	if target.After(from) {
		return target
	}

	// 下个月
	nextMonth := from.Month() + 1
	nextYear := from.Year()
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}

	// 处理日期不存在的情况（如2月30日）
	lastDay := time.Date(nextYear, nextMonth+1, 0, 0, 0, 0, 0, from.Location()).Day()
	if resetDay > lastDay {
		resetDay = lastDay
	}

	return time.Date(nextYear, nextMonth, resetDay, resetHour, resetMin, resetSec, 0, from.Location())
}

// getNextYearFirstDay 获取下年1月1号
func (s *trafficResetService) getNextYearFirstDay(from time.Time) time.Time {
	return time.Date(from.Year()+1, 1, 1, 0, 0, 0, 0, from.Location())
}

// getNextYearlyReset 按年重置（按开通日）
func (s *trafficResetService) getNextYearlyReset(user *model.User, from time.Time) time.Time {
	if user.ExpiredAt == nil {
		return s.getNextYearFirstDay(from)
	}

	expiredAt := *user.ExpiredAt
	resetMonth := expiredAt.Month()
	resetDay := expiredAt.Day()
	resetHour := expiredAt.Hour()
	resetMin := expiredAt.Minute()
	resetSec := expiredAt.Second()

	// 尝试今年
	target := time.Date(from.Year(), resetMonth, resetDay, resetHour, resetMin, resetSec, 0, from.Location())
	if target.After(from) {
		return target
	}

	// 明年
	nextYear := from.Year() + 1
	// 处理2月29日闰年
	lastDay := time.Date(nextYear, resetMonth+1, 0, 0, 0, 0, 0, from.Location()).Day()
	if resetDay > lastDay {
		resetDay = lastDay
	}

	return time.Date(nextYear, resetMonth, resetDay, resetHour, resetMin, resetSec, 0, from.Location())
}

// BatchCheckReset 批量检查并重置
func (s *trafficResetService) BatchCheckReset(batchSize int) (int, int, error) {
	if batchSize <= 0 {
		batchSize = 100
	}

	totalProcessed := 0
	totalReset := 0
	lastID := uint(0)

	for {
		var users []model.User
		err := s.db.Where("id > ? AND next_reset_at IS NOT NULL AND next_reset_at <= ? AND status = ?",
			lastID, time.Now(), 1).
			Where("expired_at IS NULL OR expired_at > ?", time.Now()).
			Order("id ASC").
			Limit(batchSize).
			Find(&users).Error

		if err != nil {
			return totalProcessed, totalReset, fmt.Errorf("failed to query users: %w", err)
		}

		if len(users) == 0 {
			break
		}

		for i := range users {
			if s.CheckAndReset(&users[i], model.TriggerSourceCron) {
				totalReset++
			}
			totalProcessed++
			lastID = users[i].ID
		}
	}

	return totalProcessed, totalReset, nil
}

// SetInitialResetTime 设置初始重置时间
func (s *trafficResetService) SetInitialResetTime(user *model.User, plan *model.Plan) {
	if plan == nil {
		return
	}

	nextReset := s.CalculateNextResetTime(user, plan)
	if nextReset != nil {
		s.db.Model(user).Update("next_reset_at", *nextReset)
	}
}

// ManualReset 管理员手动重置
func (s *trafficResetService) ManualReset(user *model.User) bool {
	return s.PerformReset(user, model.TriggerSourceManual)
}

// getResetTypeFromPlan 从套餐获取重置类型
func (s *trafficResetService) getResetTypeFromPlan(plan *model.Plan) string {
	if plan == nil {
		return model.ResetTypeManual
	}

	method := plan.ResetTrafficMethod
	if method == model.ResetTrafficFollowSystem {
		method = s.getSystemResetMethod()
	}

	switch method {
	case model.ResetTrafficFirstDayMonth:
		return model.ResetTypeFirstDayMonth
	case model.ResetTrafficMonthly:
		return model.ResetTypeMonthly
	case model.ResetTrafficFirstDayYear:
		return model.ResetTypeFirstDayYear
	case model.ResetTrafficYearly:
		return model.ResetTypeYearly
	default:
		return model.ResetTypeManual
	}
}

// getSystemResetMethod 从 settings 表获取系统默认重置方式
func (s *trafficResetService) getSystemResetMethod() int {
	var setting model.Setting
	if err := s.db.Where("`key` = ?", "reset_traffic_method").First(&setting).Error; err != nil {
		return model.ResetTrafficMonthly // 默认按月重置
	}

	var method int
	if _, err := fmt.Sscanf(setting.Value, "%d", &method); err != nil {
		return model.ResetTrafficMonthly
	}
	return method
}

// FormatTraffic 格式化流量显示
func FormatTraffic(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// toJSON helper
func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
