package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// TaskFunc 定时任务函数类型
type TaskFunc func(ctx context.Context) error

// Task 定时任务定义
type Task struct {
	Name    string
	Spec    string // cron表达式
	Func    TaskFunc
	Enabled bool
}

// Scheduler 定时任务调度器
type Scheduler struct {
	cron   *cron.Cron
	tasks  map[string]Task
	logger *zap.Logger
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewScheduler 创建调度器
func NewScheduler(logger *zap.Logger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		cron:   cron.New(cron.WithSeconds()),
		tasks:  make(map[string]Task),
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Register 注册定时任务
func (s *Scheduler) Register(task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.Name]; exists {
		return fmt.Errorf("task %s already registered", task.Name)
	}

	s.tasks[task.Name] = task
	return nil
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for name, task := range s.tasks {
		if !task.Enabled {
			s.logger.Info("Task disabled, skipping", zap.String("task", name))
			continue
		}

		taskName := name
		taskFunc := task.Func

		_, err := s.cron.AddFunc(task.Spec, func() {
			s.logger.Info("Running task", zap.String("task", taskName))
			start := time.Now()

			if err := taskFunc(s.ctx); err != nil {
				s.logger.Error("Task failed",
					zap.String("task", taskName),
					zap.Error(err),
					zap.Duration("duration", time.Since(start)),
				)
			} else {
				s.logger.Info("Task completed",
					zap.String("task", taskName),
					zap.Duration("duration", time.Since(start)),
				)
			}
		})

		if err != nil {
			return fmt.Errorf("failed to add task %s: %w", name, err)
		}
	}

	s.cron.Start()
	s.logger.Info("Scheduler started", zap.Int("tasks", len(s.tasks)))
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.cancel()
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("Scheduler stopped")
}

// GetTaskStatus 获取任务状态
func (s *Scheduler) GetTaskStatus() map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make(map[string]bool)
	for name, task := range s.tasks {
		status[name] = task.Enabled
	}
	return status
}

// EnableTask 启用任务
func (s *Scheduler) EnableTask(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[name]
	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	task.Enabled = true
	s.tasks[name] = task
	return nil
}

// DisableTask 禁用任务
func (s *Scheduler) DisableTask(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[name]
	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	task.Enabled = false
	s.tasks[name] = task
	return nil
}
