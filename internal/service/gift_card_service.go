package service

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// GiftCardService 礼品卡服务接口
type GiftCardService interface {
	// 模板管理
	CreateTemplate(name, description string, cardType int, value float64, traffic int64, duration int, planID *uint, price float64) (*model.GiftCardTemplate, error)
	GetTemplateByID(id uint) (*model.GiftCardTemplate, error)
	UpdateTemplate(id uint, name, description string, cardType int, value float64, traffic int64, duration int, planID *uint, price float64, status int) (*model.GiftCardTemplate, error)
	DeleteTemplate(id uint) error
	ListTemplates(page, pageSize int) ([]model.GiftCardTemplate, int64, error)

	// 礼品码管理
	GenerateCodes(templateID uint, count int) ([]model.GiftCardCode, error)
	GetCodeByID(id uint) (*model.GiftCardCode, error)
	GetCodeByCode(code string) (*model.GiftCardCode, error)
	DeleteCode(id uint) error
	ListCodes(page, pageSize int, templateID uint, status int) ([]model.GiftCardCode, int64, error)

	// 使用礼品卡
	UseCode(code string, userID uint) (*model.GiftCardUsage, error)
}

type giftCardService struct {
	templateRepo repository.GiftCardTemplateRepository
	codeRepo     repository.GiftCardCodeRepository
	usageRepo    repository.GiftCardUsageRepository
}

// NewGiftCardService 创建礼品卡服务
func NewGiftCardService(
	templateRepo repository.GiftCardTemplateRepository,
	codeRepo repository.GiftCardCodeRepository,
	usageRepo repository.GiftCardUsageRepository,
) GiftCardService {
	return &giftCardService{
		templateRepo: templateRepo,
		codeRepo:     codeRepo,
		usageRepo:    usageRepo,
	}
}

// CreateTemplate 创建礼品卡模板
func (s *giftCardService) CreateTemplate(name, description string, cardType int, value float64, traffic int64, duration int, planID *uint, price float64) (*model.GiftCardTemplate, error) {
	template := &model.GiftCardTemplate{
		Name:        name,
		Description: description,
		Type:        cardType,
		Value:       value,
		Traffic:     traffic,
		Duration:    duration,
		PlanID:      planID,
		Price:       price,
		Status:      1,
	}

	if err := s.templateRepo.Create(template); err != nil {
		return nil, err
	}

	return template, nil
}

// GetTemplateByID 根据ID获取礼品卡模板
func (s *giftCardService) GetTemplateByID(id uint) (*model.GiftCardTemplate, error) {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("template not found")
	}
	return template, nil
}

// UpdateTemplate 更新礼品卡模板
func (s *giftCardService) UpdateTemplate(id uint, name, description string, cardType int, value float64, traffic int64, duration int, planID *uint, price float64, status int) (*model.GiftCardTemplate, error) {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("template not found")
	}

	template.Name = name
	template.Description = description
	template.Type = cardType
	template.Value = value
	template.Traffic = traffic
	template.Duration = duration
	template.PlanID = planID
	template.Price = price
	template.Status = status

	if err := s.templateRepo.Update(template); err != nil {
		return nil, err
	}

	return template, nil
}

// DeleteTemplate 删除礼品卡模板
func (s *giftCardService) DeleteTemplate(id uint) error {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return err
	}
	if template == nil {
		return errors.New("template not found")
	}

	return s.templateRepo.Delete(id)
}

// ListTemplates 礼品卡模板列表
func (s *giftCardService) ListTemplates(page, pageSize int) ([]model.GiftCardTemplate, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.templateRepo.List(page, pageSize)
}

// GenerateCodes 批量生成礼品码
func (s *giftCardService) GenerateCodes(templateID uint, count int) ([]model.GiftCardCode, error) {
	// 验证模板是否存在
	template, err := s.templateRepo.GetByID(templateID)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("template not found")
	}

	if count <= 0 || count > 1000 {
		return nil, errors.New("count must be between 1 and 1000")
	}

	// 批量生成礼品码
	codes := make([]model.GiftCardCode, count)
	for i := 0; i < count; i++ {
		codes[i] = model.GiftCardCode{
			TemplateID: templateID,
			Code:       generateGiftCardCode(),
			Status:     0,
		}
	}

	if err := s.codeRepo.CreateBatch(codes); err != nil {
		return nil, err
	}

	return codes, nil
}

// generateGiftCardCode 生成礼品码
func generateGiftCardCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var code strings.Builder
	code.WriteString("GC")

	for i := 0; i < 16; i++ {
		if i > 0 && i%4 == 0 {
			code.WriteString("-")
		}
		code.WriteByte(chars[r.Intn(len(chars))])
	}

	return code.String()
}

// GetCodeByID 根据ID获取礼品码
func (s *giftCardService) GetCodeByID(id uint) (*model.GiftCardCode, error) {
	code, err := s.codeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if code == nil {
		return nil, errors.New("code not found")
	}
	return code, nil
}

// GetCodeByCode 根据礼品码获取
func (s *giftCardService) GetCodeByCode(code string) (*model.GiftCardCode, error) {
	giftCard, err := s.codeRepo.GetByCode(strings.ToUpper(code))
	if err != nil {
		return nil, err
	}
	if giftCard == nil {
		return nil, errors.New("code not found")
	}
	return giftCard, nil
}

// DeleteCode 删除礼品码
func (s *giftCardService) DeleteCode(id uint) error {
	code, err := s.codeRepo.GetByID(id)
	if err != nil {
		return err
	}
	if code == nil {
		return errors.New("code not found")
	}

	return s.codeRepo.Delete(id)
}

// ListCodes 礼品码列表
func (s *giftCardService) ListCodes(page, pageSize int, templateID uint, status int) ([]model.GiftCardCode, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.codeRepo.List(page, pageSize, templateID, status)
}

// UseCode 使用礼品码
func (s *giftCardService) UseCode(code string, userID uint) (*model.GiftCardUsage, error) {
	// 获取礼品码
	giftCard, err := s.codeRepo.GetByCode(strings.ToUpper(code))
	if err != nil {
		return nil, err
	}
	if giftCard == nil {
		return nil, errors.New("code not found")
	}

	// 检查是否已使用
	if giftCard.IsUsed() {
		return nil, errors.New("code already used")
	}

	// 获取模板
	template, err := s.templateRepo.GetByID(giftCard.TemplateID)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("template not found")
	}

	// 更新礼品码状态
	now := time.Now()
	giftCard.Status = 1
	giftCard.UserID = &userID
	giftCard.UsedAt = &now

	if err := s.codeRepo.Update(giftCard); err != nil {
		return nil, err
	}

	// 创建使用记录
	usage := &model.GiftCardUsage{
		CodeID:   giftCard.ID,
		UserID:   userID,
		Amount:   template.Value,
		Traffic:  template.Traffic,
		Duration: template.Duration,
	}

	if err := s.usageRepo.Create(usage); err != nil {
		return nil, fmt.Errorf("failed to create usage record: %w", err)
	}

	return usage, nil
}
