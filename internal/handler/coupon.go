package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"xboard-go/internal/service"
	"xboard-go/pkg/response"
)

// CouponHandler 优惠券处理器
type CouponHandler struct {
	couponService service.CouponService
}

// NewCouponHandler 创建优惠券处理器
func NewCouponHandler(couponService service.CouponService) *CouponHandler {
	return &CouponHandler{
		couponService: couponService,
	}
}

// CreateCouponRequest 创建优惠券请求
type CreateCouponRequest struct {
	Code        string     `json:"code" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Type        int        `json:"type" binding:"required"`
	Value       float64    `json:"value" binding:"required"`
	MinAmount   float64    `json:"min_amount"`
	MaxDiscount float64    `json:"max_discount"`
	PlanIDs     string     `json:"plan_ids"`
	UserIDs     string     `json:"user_ids"`
	LimitCount  int        `json:"limit_count"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

// CreateCoupon 创建优惠券（管理员）
func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	var req CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	coupon, err := h.couponService.Create(
		req.Code, req.Name, req.Type, req.Value,
		req.MinAmount, req.MaxDiscount, req.PlanIDs, req.UserIDs,
		req.LimitCount, req.StartDate, req.EndDate,
	)
	if err != nil {
		response.InternalError(c, "创建优惠券失败："+err.Error())
		return
	}

	response.Success(c, coupon)
}

// GetCoupon 获取优惠券详情（管理员）
func (h *CouponHandler) GetCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的优惠券ID")
		return
	}

	coupon, err := h.couponService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "优惠券不存在")
		return
	}

	response.Success(c, coupon)
}

// UpdateCouponRequest 更新优惠券请求
type UpdateCouponRequest struct {
	Code        string     `json:"code" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Type        int        `json:"type" binding:"required"`
	Value       float64    `json:"value" binding:"required"`
	MinAmount   float64    `json:"min_amount"`
	MaxDiscount float64    `json:"max_discount"`
	PlanIDs     string     `json:"plan_ids"`
	UserIDs     string     `json:"user_ids"`
	LimitCount  int        `json:"limit_count"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Status      int        `json:"status"`
}

// UpdateCoupon 更新优惠券（管理员）
func (h *CouponHandler) UpdateCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的优惠券ID")
		return
	}

	var req UpdateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	coupon, err := h.couponService.Update(
		uint(id), req.Code, req.Name, req.Type, req.Value,
		req.MinAmount, req.MaxDiscount, req.PlanIDs, req.UserIDs,
		req.LimitCount, req.StartDate, req.EndDate, req.Status,
	)
	if err != nil {
		response.InternalError(c, "更新优惠券失败："+err.Error())
		return
	}

	response.Success(c, coupon)
}

// DeleteCoupon 删除优惠券（管理员）
func (h *CouponHandler) DeleteCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的优惠券ID")
		return
	}

	if err := h.couponService.Delete(uint(id)); err != nil {
		response.InternalError(c, "删除优惠券失败")
		return
	}

	response.Success(c, nil)
}

// ListCoupons 优惠券列表（管理员）
func (h *CouponHandler) ListCoupons(c *gin.Context) {
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

	coupons, total, err := h.couponService.List(page, pageSize)
	if err != nil {
		response.InternalError(c, "获取优惠券列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  coupons,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// ValidateCouponRequest 验证优惠券请求
type ValidateCouponRequest struct {
	Code   string  `json:"code" binding:"required"`
	PlanID uint    `json:"plan_id" binding:"required"`
	Amount float64 `json:"amount" binding:"required"`
}

// ValidateCoupon 验证优惠券（用户端）
func (h *CouponHandler) ValidateCoupon(c *gin.Context) {
	var req ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	discount, err := h.couponService.Validate(req.Code, userID.(uint), req.PlanID, req.Amount)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"discount": discount,
		"total":    req.Amount - discount,
	})
}
