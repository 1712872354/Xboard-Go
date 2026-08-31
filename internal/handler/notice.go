package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// NoticeHandler 公告处理器
type NoticeHandler struct {
	noticeService service.NoticeService
}

// NewNoticeHandler 创建公告处理器
func NewNoticeHandler(noticeService service.NoticeService) *NoticeHandler {
	return &NoticeHandler{
		noticeService: noticeService,
	}
}

// CreateNoticeRequest 创建公告请求
type CreateNoticeRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	ImgURL  string `json:"img_url"`
	Show    int    `json:"show"`
	Sort    int    `json:"sort"`
	Groups  string `json:"groups"`
}

// CreateNotice 创建公告（管理员）
func (h *NoticeHandler) CreateNotice(c *gin.Context) {
	var req CreateNoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	notice, err := h.noticeService.Create(req.Title, req.Content, req.ImgURL, req.Show, req.Sort, req.Groups)
	if err != nil {
		response.InternalError(c, "创建公告失败")
		return
	}

	response.Success(c, notice)
}

// GetNotice 获取公告详情（管理员）
func (h *NoticeHandler) GetNotice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的公告ID")
		return
	}

	notice, err := h.noticeService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "公告不存在")
		return
	}

	response.Success(c, notice)
}

// UpdateNoticeRequest 更新公告请求
type UpdateNoticeRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	ImgURL  string `json:"img_url"`
	Show    int    `json:"show"`
	Sort    int    `json:"sort"`
	Groups  string `json:"groups"`
}

// UpdateNotice 更新公告（管理员）
func (h *NoticeHandler) UpdateNotice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的公告ID")
		return
	}

	var req UpdateNoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	notice, err := h.noticeService.Update(uint(id), req.Title, req.Content, req.ImgURL, req.Show, req.Sort, req.Groups)
	if err != nil {
		response.InternalError(c, "更新公告失败")
		return
	}

	response.Success(c, notice)
}

// DeleteNotice 删除公告（管理员）
func (h *NoticeHandler) DeleteNotice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的公告ID")
		return
	}

	if err := h.noticeService.Delete(uint(id)); err != nil {
		response.InternalError(c, "删除公告失败")
		return
	}

	response.Success(c, nil)
}

// ListNotices 公告列表（管理员）
func (h *NoticeHandler) ListNotices(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	notices, total, err := h.noticeService.List(page, pageSize)
	if err != nil {
		response.InternalError(c, "获取公告列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  notices,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// ListVisibleNotices 获取可见公告列表（用户端）
func (h *NoticeHandler) ListVisibleNotices(c *gin.Context) {
	notices, err := h.noticeService.ListVisible()
	if err != nil {
		response.InternalError(c, "获取公告列表失败")
		return
	}

	response.Success(c, notices)
}
