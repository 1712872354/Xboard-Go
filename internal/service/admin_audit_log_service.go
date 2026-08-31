package service

import (
	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// AdminAuditLogService 管理员审计日志服务接口
type AdminAuditLogService interface {
	Create(userID uint, username, action, resource string, resourceID uint, detail, ip, userAgent string) (*model.AdminAuditLog, error)
	GetByID(id uint) (*model.AdminAuditLog, error)
	List(page, pageSize int, userID uint, action string) ([]model.AdminAuditLog, int64, error)
	Delete(id uint) error
}

type adminAuditLogService struct {
	auditLogRepo repository.AdminAuditLogRepository
}

// NewAdminAuditLogService 创建管理员审计日志服务
func NewAdminAuditLogService(auditLogRepo repository.AdminAuditLogRepository) AdminAuditLogService {
	return &adminAuditLogService{
		auditLogRepo: auditLogRepo,
	}
}

// Create 创建审计日志
func (s *adminAuditLogService) Create(userID uint, username, action, resource string, resourceID uint, detail, ip, userAgent string) (*model.AdminAuditLog, error) {
	log := &model.AdminAuditLog{
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Detail:     detail,
		IP:         ip,
		UserAgent:  userAgent,
	}

	if err := s.auditLogRepo.Create(log); err != nil {
		return nil, err
	}

	return log, nil
}

// GetByID 根据ID获取审计日志
func (s *adminAuditLogService) GetByID(id uint) (*model.AdminAuditLog, error) {
	return s.auditLogRepo.GetByID(id)
}

// List 审计日志列表
func (s *adminAuditLogService) List(page, pageSize int, userID uint, action string) ([]model.AdminAuditLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.auditLogRepo.List(page, pageSize, userID, action)
}

// Delete 删除审计日志
func (s *adminAuditLogService) Delete(id uint) error {
	return s.auditLogRepo.Delete(id)
}
