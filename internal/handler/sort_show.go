package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"xboard-go/pkg/database"
	"xboard-go/pkg/response"

	"gorm.io/gorm"
)

// SortHandler 排序处理器
type SortHandler struct {
	db *gorm.DB
}

// NewSortHandler 创建排序处理器
func NewSortHandler() *SortHandler {
	return &SortHandler{
		db: database.Get(),
	}
}

// SortRequest 排序请求
type SortRequest struct {
	Items []SortItem `json:"items" binding:"required,min=1"`
}

// SortItem 排序项
type SortItem struct {
	ID   uint `json:"id" binding:"required"`
	Sort int  `json:"sort"`
}

// SortNotices 公告排序
// @Summary 公告排序
// @Description 批量更新公告排序
// @Tags 管理员-公告
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body SortRequest true "排序信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/notices/sort [post]
func (h *SortHandler) SortNotices(c *gin.Context) {
	h.handleSort(c, "notices")
}

// SortPlans 套餐排序
// @Summary 套餐排序
// @Description 批量更新套餐排序
// @Tags 管理员-套餐
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body SortRequest true "排序信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/plans/sort [post]
func (h *SortHandler) SortPlans(c *gin.Context) {
	h.handleSort(c, "plans")
}

// SortKnowledges 知识库排序
// @Summary 知识库排序
// @Description 批量更新知识库排序
// @Tags 管理员-知识库
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body SortRequest true "排序信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/knowledges/sort [post]
func (h *SortHandler) SortKnowledges(c *gin.Context) {
	h.handleSort(c, "knowledges")
}

// SortPayments 支付方式排序
// @Summary 支付方式排序
// @Description 批量更新支付方式排序
// @Tags 管理员-支付
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body SortRequest true "排序信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/payments/sort [post]
func (h *SortHandler) SortPayments(c *gin.Context) {
	h.handleSort(c, "payments")
}

// handleSort 处理排序
func (h *SortHandler) handleSort(c *gin.Context, table string) {
	var req SortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	tx := h.db.Begin()
	for _, item := range req.Items {
		if err := tx.Table(table).Where("id = ?", item.ID).Update("sort", item.Sort).Error; err != nil {
			tx.Rollback()
			response.InternalError(c, "排序失败")
			return
		}
	}
	tx.Commit()

	response.Success(c, nil)
}

// ShowHandler 显示状态处理器
type ShowHandler struct {
	db *gorm.DB
}

// NewShowHandler 创建显示状态处理器
func NewShowHandler() *ShowHandler {
	return &ShowHandler{
		db: database.Get(),
	}
}

// ShowRequest 显示状态请求
type ShowRequest struct {
	Show int `json:"show" binding:"required,min=0,max=1"`
}

// UpdateNoticeShow 更新公告显示状态
// @Summary 更新公告显示状态
// @Description 更新公告的显示/隐藏状态
// @Tags 管理员-公告
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "公告ID"
// @Param request body ShowRequest true "显示状态"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/notices/{id}/show [put]
func (h *ShowHandler) UpdateNoticeShow(c *gin.Context) {
	h.handleShow(c, "notices")
}

// UpdateKnowledgeShow 更新知识库显示状态
// @Summary 更新知识库显示状态
// @Description 更新知识库的显示/隐藏状态
// @Tags 管理员-知识库
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "知识库ID"
// @Param request body ShowRequest true "显示状态"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/knowledges/{id}/show [put]
func (h *ShowHandler) UpdateKnowledgeShow(c *gin.Context) {
	h.handleShow(c, "knowledges")
}

// UpdateCouponShow 更新优惠券显示状态
// @Summary 更新优惠券显示状态
// @Description 更新优惠券的显示/隐藏状态
// @Tags 管理员-优惠券
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "优惠券ID"
// @Param request body ShowRequest true "显示状态"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/coupons/{id}/show [put]
func (h *ShowHandler) UpdateCouponShow(c *gin.Context) {
	h.handleShow(c, "coupons")
}

// handleShow 处理显示状态更新
func (h *ShowHandler) handleShow(c *gin.Context, table string) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req ShowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.db.Table(table).Where("id = ?", id).Update("show", req.Show).Error; err != nil {
		response.InternalError(c, "更新显示状态失败")
		return
	}

	response.Success(c, nil)
}
