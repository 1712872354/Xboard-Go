package service

import (
	"errors"

	"xboard-go/internal/model"
	"xboard-go/internal/repository"
)

// NoticeService 公告服务接口
type NoticeService interface {
	Create(title, content, imgURL string, show, sort int, groups string) (*model.Notice, error)
	GetByID(id uint) (*model.Notice, error)
	Update(id uint, title, content, imgURL string, show, sort int, groups string) (*model.Notice, error)
	Delete(id uint) error
	List(page, pageSize int) ([]model.Notice, int64, error)
	ListVisible() ([]model.Notice, error)
}

type noticeService struct {
	noticeRepo repository.NoticeRepository
}

// NewNoticeService 创建公告服务
func NewNoticeService(noticeRepo repository.NoticeRepository) NoticeService {
	return &noticeService{
		noticeRepo: noticeRepo,
	}
}

// Create 创建公告
func (s *noticeService) Create(title, content, imgURL string, show, sort int, groups string) (*model.Notice, error) {
	notice := &model.Notice{
		Title:   title,
		Content: content,
		ImgURL:  imgURL,
		Show:    show,
		Sort:    sort,
		Groups:  groups,
	}

	if err := s.noticeRepo.Create(notice); err != nil {
		return nil, err
	}

	return notice, nil
}

// GetByID 根据ID获取公告
func (s *noticeService) GetByID(id uint) (*model.Notice, error) {
	notice, err := s.noticeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if notice == nil {
		return nil, errors.New("notice not found")
	}
	return notice, nil
}

// Update 更新公告
func (s *noticeService) Update(id uint, title, content, imgURL string, show, sort int, groups string) (*model.Notice, error) {
	notice, err := s.noticeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if notice == nil {
		return nil, errors.New("notice not found")
	}

	notice.Title = title
	notice.Content = content
	notice.ImgURL = imgURL
	notice.Show = show
	notice.Sort = sort
	notice.Groups = groups

	if err := s.noticeRepo.Update(notice); err != nil {
		return nil, err
	}

	return notice, nil
}

// Delete 删除公告
func (s *noticeService) Delete(id uint) error {
	notice, err := s.noticeRepo.GetByID(id)
	if err != nil {
		return err
	}
	if notice == nil {
		return errors.New("notice not found")
	}

	return s.noticeRepo.Delete(id)
}

// List 公告列表（管理员）
func (s *noticeService) List(page, pageSize int) ([]model.Notice, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.noticeRepo.List(page, pageSize)
}

// ListVisible 获取可见公告列表（用户端）
func (s *noticeService) ListVisible() ([]model.Notice, error) {
	return s.noticeRepo.ListVisible()
}
