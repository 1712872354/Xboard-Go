package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
	"xboard-go/pkg/database"
	"xboard-go/pkg/logger"
	appredis "xboard-go/pkg/redis"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	devicePrefix = "user_devices:"
	deviceTTL    = 300 // 5分钟 TTL
)

// DeviceService 设备状态服务接口
type DeviceService interface {
	// SetDevices 批量设置用户设备（节点上报）
	SetDevices(userID uint, nodeID uint, ips []string) error
	// GetDeviceCount 获取用户在线设备数（去重IP）
	GetDeviceCount(userID uint) (int, error)
	// GetUserDevices 获取用户设备详情
	GetUserDevices(userID uint) (map[string][]string, error)
	// GetNodeDevices 获取节点所有设备
	GetNodeDevices(nodeID uint) (map[uint][]string, error)
	// ClearNodeDevices 清除节点所有设备
	ClearNodeDevices(nodeID uint) error
	// CheckDeviceLimit 检查设备限制
	CheckDeviceLimit(userID uint, limit int) (bool, error)
	// CleanupOfflineDevices 清理离线设备
	CleanupOfflineDevices() error
	// GetAliveList 获取在线设备列表（用于 alivelist 接口）
	GetAliveList(userIDs []uint) map[uint]int
}

// deviceService 设备状态服务实现
type deviceService struct {
	userRepo repository.UserRepository
	nodeRepo repository.NodeRepository
	db       *gorm.DB
}

// NewDeviceService 创建设备状态服务
func NewDeviceService(userRepo repository.UserRepository, nodeRepo repository.NodeRepository) DeviceService {
	return &deviceService{
		userRepo: userRepo,
		nodeRepo: nodeRepo,
		db:       database.Get(),
	}
}

// SetDevices 批量设置用户设备
func (s *deviceService) SetDevices(userID uint, nodeID uint, ips []string) error {
	client := appredis.Client()
	if client == nil {
		return nil
	}

	ctx := context.Background()
	key := fmt.Sprintf("%s%d", devicePrefix, userID)

	// 先删除该节点的旧设备数据
	s.removeNodeDevices(ctx, client, userID, nodeID)

	if len(ips) == 0 {
		return nil
	}

	// 去重 IP（移除端口）
	normalizedIPs := normalizeIPs(ips)
	now := time.Now().Unix()

	// 使用 Hash 存储：field = "nodeID:IP", value = timestamp
	fields := make(map[string]interface{})
	for _, ip := range normalizedIPs {
		field := fmt.Sprintf("%d:%s", nodeID, ip)
		fields[field] = now
	}

	if err := client.HSet(ctx, key, fields).Err(); err != nil {
		return fmt.Errorf("failed to set devices: %w", err)
	}

	// 设置 TTL
	client.Expire(ctx, key, deviceTTL*time.Second)

	// 更新用户在线设备数
	count := s.getDeviceCountFromRedis(ctx, client, userID)
	s.updateUserOnlineCount(userID, count)

	return nil
}

// GetDeviceCount 获取用户在线设备数
func (s *deviceService) GetDeviceCount(userID uint) (int, error) {
	client := appredis.Client()
	if client == nil {
		return 0, nil
	}

	ctx := context.Background()
	return s.getDeviceCountFromRedis(ctx, client, userID), nil
}

// getDeviceCountFromRedis 从 Redis 获取设备数
func (s *deviceService) getDeviceCountFromRedis(ctx context.Context, client *redis.Client, userID uint) int {
	key := fmt.Sprintf("%s%d", devicePrefix, userID)
	data, err := client.HGetAll(ctx, key).Result()
	if err != nil {
		return 0
	}

	now := time.Now().Unix()
	ips := make(map[string]bool)

	for field, timestampStr := range data {
		var timestamp int64
		fmt.Sscanf(timestampStr, "%d", &timestamp)

		// 检查是否过期
		if now-timestamp > deviceTTL {
			continue
		}

		// 提取 IP（去掉 nodeID: 前缀）
		parts := strings.SplitN(field, ":", 2)
		if len(parts) == 2 {
			ips[parts[1]] = true
		}
	}

	return len(ips)
}

// GetUserDevices 获取用户设备详情
func (s *deviceService) GetUserDevices(userID uint) (map[string][]string, error) {
	client := appredis.Client()
	if client == nil {
		return nil, nil
	}

	ctx := context.Background()
	key := fmt.Sprintf("%s%d", devicePrefix, userID)
	data, err := client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	result := make(map[string][]string) // nodeID -> IPs

	for field, timestampStr := range data {
		var timestamp int64
		fmt.Sscanf(timestampStr, "%d", &timestamp)

		if now-timestamp > deviceTTL {
			continue
		}

		parts := strings.SplitN(field, ":", 2)
		if len(parts) == 2 {
			nodeID := parts[0]
			ip := parts[1]
			result[nodeID] = append(result[nodeID], ip)
		}
	}

	return result, nil
}

// GetNodeDevices 获取节点所有设备
func (s *deviceService) GetNodeDevices(nodeID uint) (map[uint][]string, error) {
	client := appredis.Client()
	if client == nil {
		return nil, nil
	}

	ctx := context.Background()
	pattern := fmt.Sprintf("%s*", devicePrefix)
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	result := make(map[uint][]string) // userID -> IPs
	prefix := fmt.Sprintf("%d:", nodeID)

	for _, key := range keys {
		data, err := client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		// 提取 userID
		var userID uint
		fmt.Sscanf(key, devicePrefix+"%d", &userID)

		for field, timestampStr := range data {
			var timestamp int64
			fmt.Sscanf(timestampStr, "%d", &timestamp)

			if now-timestamp > deviceTTL {
				continue
			}

			if strings.HasPrefix(field, prefix) {
				ip := field[len(prefix):]
				result[userID] = append(result[userID], ip)
			}
		}
	}

	return result, nil
}

// ClearNodeDevices 清除节点所有设备
func (s *deviceService) ClearNodeDevices(nodeID uint) error {
	client := appredis.Client()
	if client == nil {
		return nil
	}

	ctx := context.Background()
	pattern := fmt.Sprintf("%s*", devicePrefix)
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("%d:", nodeID)

	for _, key := range keys {
		data, err := client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		for field := range data {
			if strings.HasPrefix(field, prefix) {
				client.HDel(ctx, key, field)
			}
		}
	}

	return nil
}

// CheckDeviceLimit 检查设备限制
func (s *deviceService) CheckDeviceLimit(userID uint, limit int) (bool, error) {
	if limit <= 0 {
		return true, nil // 0 表示不限制
	}

	count, err := s.GetDeviceCount(userID)
	if err != nil {
		return false, err
	}

	return count < limit, nil
}

// GetAliveList 获取在线设备列表
func (s *deviceService) GetAliveList(userIDs []uint) map[uint]int {
	result := make(map[uint]int)
	for _, userID := range userIDs {
		count, err := s.GetDeviceCount(userID)
		if err != nil {
			continue
		}
		if count > 0 {
			result[userID] = count
		}
	}
	return result
}

// CleanupOfflineDevices 清理离线设备
func (s *deviceService) CleanupOfflineDevices() error {
	client := appredis.Client()
	if client == nil {
		return nil
	}

	ctx := context.Background()
	pattern := fmt.Sprintf("%s*", devicePrefix)
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	cleaned := 0

	for _, key := range keys {
		data, err := client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		for field, timestampStr := range data {
			var timestamp int64
			fmt.Sscanf(timestampStr, "%d", &timestamp)

			if now-timestamp > deviceTTL {
				client.HDel(ctx, key, field)
				cleaned++
			}
		}
	}

	if cleaned > 0 {
		logger.Sugar().Debugf("Cleaned %d expired device entries", cleaned)
	}

	return nil
}

// removeNodeDevices 删除节点旧设备数据
func (s *deviceService) removeNodeDevices(ctx context.Context, client *redis.Client, userID uint, nodeID uint) {
	key := fmt.Sprintf("%s%d", devicePrefix, userID)
	prefix := fmt.Sprintf("%d:", nodeID)

	data, err := client.HGetAll(ctx, key).Result()
	if err != nil {
		return
	}

	for field := range data {
		if strings.HasPrefix(field, prefix) {
			client.HDel(ctx, key, field)
		}
	}
}

// updateUserOnlineCount 更新用户在线设备数
func (s *deviceService) updateUserOnlineCount(userID uint, count int) {
	if err := s.db.Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"online_count":  count,
			"last_online_at": time.Now(),
		}).Error; err != nil {
		logger.Sugar().Warnf("Failed to update online count for user %d: %v", userID, err)
	}
}

// normalizeIPs 标准化 IP 地址（移除端口）
func normalizeIPs(ips []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, ip := range ips {
		normalized := normalizeIP(ip)
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}

	return result
}

// normalizeIP 标准化单个 IP（移除端口）
func normalizeIP(ip string) string {
	// IPv6: [::1]:port
	if strings.HasPrefix(ip, "[") {
		if idx := strings.LastIndex(ip, "]:"); idx > 0 {
			return ip[1:idx]
		}
		return ip
	}

	// IPv4: 1.2.3.4:port
	if idx := strings.LastIndex(ip, ":"); idx > 0 {
		return ip[:idx]
	}

	return ip
}
