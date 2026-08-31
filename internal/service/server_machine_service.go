package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"xboard-go/config"
	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// ServerMachineService 服务器机器服务接口
type ServerMachineService interface {
	Create(name, host string, port int, protocol string) (*model.ServerMachine, error)
	GetByID(id uint) (*model.ServerMachine, error)
	Update(id uint, name, host string, port int, protocol string, status int) (*model.ServerMachine, error)
	Delete(id uint) error
	List(page, pageSize int) ([]model.ServerMachine, int64, error)
	ListAll() ([]model.ServerMachine, error)
	UpdateStatus(id uint, status int) error
	UpdateLoad(id uint, cpu, memory, disk float64) error
	ResetToken(id uint) (*model.ServerMachine, error)
	GetInstallCommand(id uint) (string, error)
}

type serverMachineService struct {
	machineRepo repository.ServerMachineRepository
}

// NewServerMachineService 创建服务器机器服务
func NewServerMachineService(machineRepo repository.ServerMachineRepository) ServerMachineService {
	return &serverMachineService{
		machineRepo: machineRepo,
	}
}

// Create 创建服务器机器
func (s *serverMachineService) Create(name, host string, port int, protocol string) (*model.ServerMachine, error) {
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	machine := &model.ServerMachine{
		Name:     name,
		Host:     host,
		Port:     port,
		Protocol: protocol,
		Token:    token,
		Status:   1,
	}

	if err := s.machineRepo.Create(machine); err != nil {
		return nil, err
	}

	return machine, nil
}

// GetByID 根据ID获取服务器机器
func (s *serverMachineService) GetByID(id uint) (*model.ServerMachine, error) {
	machine, err := s.machineRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if machine == nil {
		return nil, errors.New("server machine not found")
	}
	return machine, nil
}

// Update 更新服务器机器
func (s *serverMachineService) Update(id uint, name, host string, port int, protocol string, status int) (*model.ServerMachine, error) {
	machine, err := s.machineRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if machine == nil {
		return nil, errors.New("server machine not found")
	}

	machine.Name = name
	machine.Host = host
	machine.Port = port
	machine.Protocol = protocol
	machine.Status = status

	if err := s.machineRepo.Update(machine); err != nil {
		return nil, err
	}

	return machine, nil
}

// Delete 删除服务器机器
func (s *serverMachineService) Delete(id uint) error {
	machine, err := s.machineRepo.GetByID(id)
	if err != nil {
		return err
	}
	if machine == nil {
		return errors.New("server machine not found")
	}

	return s.machineRepo.Delete(id)
}

// List 服务器机器列表
func (s *serverMachineService) List(page, pageSize int) ([]model.ServerMachine, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.machineRepo.List(page, pageSize)
}

// ListAll 获取所有服务器机器
func (s *serverMachineService) ListAll() ([]model.ServerMachine, error) {
	return s.machineRepo.ListAll()
}

// UpdateStatus 更新服务器机器状态
func (s *serverMachineService) UpdateStatus(id uint, status int) error {
	machine, err := s.machineRepo.GetByID(id)
	if err != nil {
		return err
	}
	if machine == nil {
		return errors.New("server machine not found")
	}

	return s.machineRepo.UpdateStatus(id, status)
}

// UpdateLoad 更新服务器机器负载
func (s *serverMachineService) UpdateLoad(id uint, cpu, memory, disk float64) error {
	machine, err := s.machineRepo.GetByID(id)
	if err != nil {
		return err
	}
	if machine == nil {
		return errors.New("server machine not found")
	}

	return s.machineRepo.UpdateLoad(id, cpu, memory, disk)
}

// ResetToken 重新生成Token
func (s *serverMachineService) ResetToken(id uint) (*model.ServerMachine, error) {
	machine, err := s.machineRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if machine == nil {
		return nil, errors.New("server machine not found")
	}

	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	if err := s.machineRepo.UpdateToken(id, token); err != nil {
		return nil, err
	}

	machine.Token = token
	return machine, nil
}

// GetInstallCommand 生成安装命令
func (s *serverMachineService) GetInstallCommand(id uint) (string, error) {
	machine, err := s.machineRepo.GetByID(id)
	if err != nil {
		return "", err
	}
	if machine == nil {
		return "", errors.New("server machine not found")
	}
	if machine.Token == "" {
		return "", errors.New("machine token not set, please reset token first")
	}

	cfg := config.Get()
	panelURL := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
	cmd := fmt.Sprintf("curl -fsSL %s/install.sh | bash -s -- --panel-url %s --api-key %s --node-id %d",
		panelURL, panelURL, machine.Token, machine.ID)

	return cmd, nil
}

// generateToken 生成32字符的随机hex token
func generateToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
