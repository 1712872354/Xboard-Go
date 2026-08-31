package grpc

import (
	"sync"
	"time"
)

// CachedMetrics holds the latest system metrics reported by a node.
type CachedMetrics struct {
	CPU         float32   `json:"cpu"`
	MemTotal    uint64    `json:"mem_total"`
	MemUsed     uint64    `json:"mem_used"`
	DiskTotal   uint64    `json:"disk_total"`
	DiskUsed    uint64    `json:"disk_used"`
	Uptime      uint64    `json:"uptime"`
	Goroutines  int32     `json:"goroutines"`
	ActiveConns int32     `json:"active_conns"`
	ActiveUsers int32     `json:"active_users"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NodeMetricsCache is a thread-safe cache for node metrics.
type NodeMetricsCache struct {
	mu      sync.RWMutex
	metrics map[uint32]*CachedMetrics
}

// GlobalMetricsCache is the singleton instance used across the application.
var GlobalMetricsCache *NodeMetricsCache

func init() {
	GlobalMetricsCache = &NodeMetricsCache{
		metrics: make(map[uint32]*CachedMetrics),
	}
}

// UpdateMetrics stores the latest metrics for a given node.
func (c *NodeMetricsCache) UpdateMetrics(nodeID uint32, report *StatusReport) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics[nodeID] = &CachedMetrics{
		CPU:         report.CPU,
		MemTotal:    report.MemTotal,
		MemUsed:     report.MemUsed,
		DiskTotal:   report.DiskTotal,
		DiskUsed:    report.DiskUsed,
		Uptime:      report.Uptime,
		Goroutines:  report.Goroutines,
		ActiveConns: report.ActiveConns,
		ActiveUsers: report.ActiveUsers,
		UpdatedAt:   time.Now(),
	}
}

// GetMetrics returns the cached metrics for a given node, or nil if not available.
func (c *NodeMetricsCache) GetMetrics(nodeID uint32) *CachedMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.metrics[nodeID]
}

// GetAllMetrics returns a snapshot of all cached metrics.
func (c *NodeMetricsCache) GetAllMetrics() map[uint32]*CachedMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[uint32]*CachedMetrics, len(c.metrics))
	for k, v := range c.metrics {
		result[k] = v
	}
	return result
}
