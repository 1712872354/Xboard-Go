package service

import (
	"sync"
	"time"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// SettingService 系统设置服务接口
type SettingService interface {
	GetByKey(key string) (string, error)
	GetByGroup(group string) (map[string]string, error)
	GetAll() (map[string]map[string]string, error)
	Set(key, value, group, remark string) error
	SetBatch(settings []model.Setting) error
	Delete(key string) error

	// 便捷方法
	GetAppName() string
	GetAppURL() string
	GetSubscribePath() string
}

type settingService struct {
	settingRepo repository.SettingRepository

	// 缓存
	cache   map[string]*settingCacheItem
	cacheMu sync.RWMutex
}

type settingCacheItem struct {
	value     string
	expiresAt time.Time
}

// NewSettingService 创建系统设置服务
func NewSettingService(settingRepo repository.SettingRepository) SettingService {
	return &settingService{
		settingRepo: settingRepo,
		cache:       make(map[string]*settingCacheItem),
	}
}

// GetByKey 获取单个设置值
func (s *settingService) GetByKey(key string) (string, error) {
	// 检查缓存（10分钟有效期）
	s.cacheMu.RLock()
	if item, exists := s.cache[key]; exists && time.Now().Before(item.expiresAt) {
		defer s.cacheMu.RUnlock()
		return item.value, nil
	}
	s.cacheMu.RUnlock()

	// 缓存过期或不存在，查询数据库
	setting, err := s.settingRepo.GetByKey(key)
	if err != nil {
		return "", err
	}

	var value string
	if setting != nil {
		value = setting.Value
	}

	// 更新缓存
	s.cacheMu.Lock()
	s.cache[key] = &settingCacheItem{
		value:     value,
		expiresAt: time.Now().Add(10 * time.Minute),
	}
	s.cacheMu.Unlock()

	return value, nil
}

// GetByGroup 获取分组设置
func (s *settingService) GetByGroup(group string) (map[string]string, error) {
	settings, err := s.settingRepo.GetByGroup(group)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(settings))
	for _, setting := range settings {
		result[setting.Key] = setting.Value
	}
	return result, nil
}

// GetAll 获取所有设置（按分组）
func (s *settingService) GetAll() (map[string]map[string]string, error) {
	settings, err := s.settingRepo.GetAll()
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]string)
	for _, setting := range settings {
		if _, ok := result[setting.Group]; !ok {
			result[setting.Group] = make(map[string]string)
		}
		result[setting.Group][setting.Key] = setting.Value
	}
	return result, nil
}

// Set 设置单个配置项
func (s *settingService) Set(key, value, group, remark string) error {
	// 清除缓存
	s.cacheMu.Lock()
	delete(s.cache, key)
	s.cacheMu.Unlock()

	return s.settingRepo.Set(key, value, group, remark)
}

// SetBatch 批量设置配置项
func (s *settingService) SetBatch(settings []model.Setting) error {
	return s.settingRepo.SetBatch(settings)
}

// Delete 删除设置
func (s *settingService) Delete(key string) error {
	// 清除缓存
	s.cacheMu.Lock()
	delete(s.cache, key)
	s.cacheMu.Unlock()

	return s.settingRepo.Delete(key)
}

// GetAppName 获取应用名称
func (s *settingService) GetAppName() string {
	name, _ := s.GetByKey(model.SettingKeyAppName)
	if name == "" {
		return "Xboard"
	}
	return name
}

// GetAppURL 获取应用URL
func (s *settingService) GetAppURL() string {
	url, _ := s.GetByKey(model.SettingKeyAppURL)
	return url
}

// GetSubscribePath 获取订阅路径
func (s *settingService) GetSubscribePath() string {
	path, _ := s.GetByKey(model.SettingKeySubscribePath)
	if path == "" {
		return "s"
	}
	return path
}
