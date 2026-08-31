package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// KnowledgeHandler 知识库处理器
type KnowledgeHandler struct {
	knowledgeService service.KnowledgeService
}

// NewKnowledgeHandler 创建知识库处理器
func NewKnowledgeHandler(knowledgeService service.KnowledgeService) *KnowledgeHandler {
	return &KnowledgeHandler{
		knowledgeService: knowledgeService,
	}
}

// CreateKnowledgeRequest 创建知识库文章请求
type CreateKnowledgeRequest struct {
	Category string `json:"category" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Language string `json:"language"`
	Show     int    `json:"show"`
	Sort     int    `json:"sort"`
}

// CreateKnowledge 创建知识库文章（管理员）
func (h *KnowledgeHandler) CreateKnowledge(c *gin.Context) {
	var req CreateKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if req.Language == "" {
		req.Language = "zh-CN"
	}

	knowledge, err := h.knowledgeService.Create(req.Category, req.Title, req.Content, req.Language, req.Show, req.Sort)
	if err != nil {
		response.InternalError(c, "创建知识库文章失败")
		return
	}

	response.Success(c, knowledge)
}

// GetKnowledge 获取知识库文章详情（管理员）
func (h *KnowledgeHandler) GetKnowledge(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	knowledge, err := h.knowledgeService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "文章不存在")
		return
	}

	response.Success(c, knowledge)
}

// UpdateKnowledgeRequest 更新知识库文章请求
type UpdateKnowledgeRequest struct {
	Category string `json:"category" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Language string `json:"language"`
	Show     int    `json:"show"`
	Sort     int    `json:"sort"`
}

// UpdateKnowledge 更新知识库文章（管理员）
func (h *KnowledgeHandler) UpdateKnowledge(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	var req UpdateKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if req.Language == "" {
		req.Language = "zh-CN"
	}

	knowledge, err := h.knowledgeService.Update(uint(id), req.Category, req.Title, req.Content, req.Language, req.Show, req.Sort)
	if err != nil {
		response.InternalError(c, "更新知识库文章失败")
		return
	}

	response.Success(c, knowledge)
}

// DeleteKnowledge 删除知识库文章（管理员）
func (h *KnowledgeHandler) DeleteKnowledge(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	if err := h.knowledgeService.Delete(uint(id)); err != nil {
		response.InternalError(c, "删除知识库文章失败")
		return
	}

	response.Success(c, nil)
}

// ListKnowledges 知识库列表（管理员）
func (h *KnowledgeHandler) ListKnowledges(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	category := c.Query("category")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	knowledges, total, err := h.knowledgeService.List(page, pageSize, category)
	if err != nil {
		response.InternalError(c, "获取知识库列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  knowledges,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// ListVisibleKnowledges 获取可见知识库列表（用户端）
func (h *KnowledgeHandler) ListVisibleKnowledges(c *gin.Context) {
	category := c.Query("category")
	language := c.DefaultQuery("language", "zh-CN")

	knowledges, err := h.knowledgeService.ListVisible(category, language)
	if err != nil {
		response.InternalError(c, "获取知识库列表失败")
		return
	}

	response.Success(c, knowledges)
}

// GetCategories 获取知识库分类列表
func (h *KnowledgeHandler) GetCategories(c *gin.Context) {
	categories, err := h.knowledgeService.GetCategories()
	if err != nil {
		response.InternalError(c, "获取分类列表失败")
		return
	}

	response.Success(c, categories)
}
