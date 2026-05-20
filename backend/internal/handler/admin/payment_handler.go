package admin

import (
	"context"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentHandler handles admin payment management.
type PaymentHandler struct {
	paymentService *service.PaymentService
	configService  *service.PaymentConfigService
}

// NewPaymentHandler creates a new admin PaymentHandler.
func NewPaymentHandler(paymentService *service.PaymentService, configService *service.PaymentConfigService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		configService:  configService,
	}
}

// --- Dashboard ---

// GetDashboard returns payment dashboard statistics.
// GET /api/v1/admin/payment/dashboard
func (h *PaymentHandler) GetDashboard(c *gin.Context) {
	days := 30
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}
	stats, err := h.paymentService.GetDashboardStats(c.Request.Context(), days)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, stats)
}

// --- Orders ---

// ListOrders returns a paginated list of all payment orders.
// GET /api/v1/admin/payment/orders
func (h *PaymentHandler) ListOrders(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	var userID int64
	if uid := c.Query("user_id"); uid != "" {
		if v, err := strconv.ParseInt(uid, 10, 64); err == nil {
			userID = v
		}
	}
	orders, total, err := h.paymentService.AdminListOrders(c.Request.Context(), userID, service.OrderListParams{
		Page:        page,
		PageSize:    pageSize,
		Status:      c.Query("status"),
		OrderType:   c.Query("order_type"),
		PaymentType: c.Query("payment_type"),
		Keyword:     c.Query("keyword"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, orders, int64(total), page, pageSize)
}

// GetOrderDetail returns detailed information about a single order.
// GET /api/v1/admin/payment/orders/:id
func (h *PaymentHandler) GetOrderDetail(c *gin.Context) {
	orderID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	order, err := h.paymentService.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	auditLogs, _ := h.paymentService.GetOrderAuditLogs(c.Request.Context(), orderID)
	response.Success(c, gin.H{"order": order, "auditLogs": auditLogs})
}

// CancelOrder cancels a pending order (admin).
// POST /api/v1/admin/payment/orders/:id/cancel
func (h *PaymentHandler) CancelOrder(c *gin.Context) {
	orderID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	payload := struct {
		OrderID int64 `json:"order_id"`
	}{OrderID: orderID}
	executeAdminIdempotentJSON(c, "admin.payment.orders.cancel", payload, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		msg, err := h.paymentService.AdminCancelOrder(ctx, orderID)
		if err != nil {
			return nil, err
		}
		return gin.H{"message": msg}, nil
	})
}

// RetryFulfillment retries fulfillment for a paid order.
// POST /api/v1/admin/payment/orders/:id/retry
func (h *PaymentHandler) RetryFulfillment(c *gin.Context) {
	orderID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	payload := struct {
		OrderID int64 `json:"order_id"`
	}{OrderID: orderID}
	executeAdminIdempotentJSON(c, "admin.payment.orders.retry_fulfillment", payload, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		if err := h.paymentService.RetryFulfillment(ctx, orderID); err != nil {
			return nil, err
		}
		return gin.H{"message": "fulfillment retried"}, nil
	})
}

// AdminProcessRefundRequest is the request body for admin refund processing.
type AdminProcessRefundRequest struct {
	Amount        float64 `json:"amount"`
	Reason        string  `json:"reason"`
	Force         bool    `json:"force"`
	DeductBalance bool    `json:"deduct_balance"`
}

// ProcessRefund processes a refund for an order (admin).
// POST /api/v1/admin/payment/orders/:id/refund
func (h *PaymentHandler) ProcessRefund(c *gin.Context) {
	orderID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req AdminProcessRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	payload := struct {
		OrderID int64                     `json:"order_id"`
		Body    AdminProcessRefundRequest `json:"body"`
	}{OrderID: orderID, Body: req}
	executeAdminIdempotentJSON(c, "admin.payment.orders.process_refund", payload, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		plan, earlyResult, err := h.paymentService.PrepareRefund(ctx, orderID, req.Amount, req.Reason, req.Force, req.DeductBalance)
		if err != nil {
			return nil, err
		}
		if earlyResult != nil {
			return earlyResult, nil
		}
		return h.paymentService.ExecuteRefund(ctx, plan)
	})
}

// --- Subscription Plans ---

// ListPlans returns all subscription plans.
// GET /api/v1/admin/payment/plans
func (h *PaymentHandler) ListPlans(c *gin.Context) {
	plans, err := h.configService.ListPlans(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, plans)
}

// CreatePlan creates a new subscription plan.
// POST /api/v1/admin/payment/plans
func (h *PaymentHandler) CreatePlan(c *gin.Context) {
	var req service.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := executeAdminIdempotent(c, "admin.payment.plans.create", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		return h.configService.CreatePlan(ctx, req)
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if result != nil && result.Replayed {
		c.Header("X-Idempotency-Replayed", "true")
	}
	response.Created(c, result.Data)
}

// UpdatePlan updates an existing subscription plan.
// PUT /api/v1/admin/payment/plans/:id
func (h *PaymentHandler) UpdatePlan(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req service.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	payload := struct {
		ID   int64                     `json:"id"`
		Body service.UpdatePlanRequest `json:"body"`
	}{ID: id, Body: req}
	executeAdminIdempotentJSON(c, "admin.payment.plans.update", payload, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		return h.configService.UpdatePlan(ctx, id, req)
	})
}

// DeletePlan deletes a subscription plan.
// DELETE /api/v1/admin/payment/plans/:id
func (h *PaymentHandler) DeletePlan(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	payload := struct {
		ID int64 `json:"id"`
	}{ID: id}
	executeAdminIdempotentJSON(c, "admin.payment.plans.delete", payload, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		if err := h.configService.DeletePlan(ctx, id); err != nil {
			return nil, err
		}
		return gin.H{"message": "deleted"}, nil
	})
}

// --- Provider Instances ---

// ListProviders returns all payment provider instances.
// GET /api/v1/admin/payment/providers
func (h *PaymentHandler) ListProviders(c *gin.Context) {
	providers, err := h.configService.ListProviderInstancesWithConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, providers)
}

// CreateProvider creates a new payment provider instance.
// POST /api/v1/admin/payment/providers
func (h *PaymentHandler) CreateProvider(c *gin.Context) {
	var req service.CreateProviderInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := executeAdminIdempotent(c, "admin.payment.providers.create", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		inst, err := h.configService.CreateProviderInstance(ctx, req)
		if err != nil {
			return nil, err
		}
		h.paymentService.RefreshProviders(ctx)
		return inst, nil
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if result != nil && result.Replayed {
		c.Header("X-Idempotency-Replayed", "true")
	}
	response.Created(c, result.Data)
}

// UpdateProvider updates an existing payment provider instance.
// PUT /api/v1/admin/payment/providers/:id
func (h *PaymentHandler) UpdateProvider(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req service.UpdateProviderInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	payload := struct {
		ID   int64                                 `json:"id"`
		Body service.UpdateProviderInstanceRequest `json:"body"`
	}{ID: id, Body: req}
	executeAdminIdempotentJSON(c, "admin.payment.providers.update", payload, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		inst, err := h.configService.UpdateProviderInstance(ctx, id, req)
		if err != nil {
			return nil, err
		}
		h.paymentService.RefreshProviders(ctx)
		return inst, nil
	})
}

// DeleteProvider deletes a payment provider instance.
// DELETE /api/v1/admin/payment/providers/:id
func (h *PaymentHandler) DeleteProvider(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	payload := struct {
		ID int64 `json:"id"`
	}{ID: id}
	executeAdminIdempotentJSON(c, "admin.payment.providers.delete", payload, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		if err := h.configService.DeleteProviderInstance(ctx, id); err != nil {
			return nil, err
		}
		h.paymentService.RefreshProviders(ctx)
		return gin.H{"message": "deleted"}, nil
	})
}

// parseIDParam parses an int64 path parameter.
// Returns the parsed ID and true on success; on failure it writes a BadRequest response and returns false.
func parseIDParam(c *gin.Context, paramName string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(paramName), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid "+paramName)
		return 0, false
	}
	return id, true
}

// --- Config ---

// GetConfig returns the payment configuration (admin view).
// GET /api/v1/admin/payment/config
func (h *PaymentHandler) GetConfig(c *gin.Context) {
	cfg, err := h.configService.GetPaymentConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// UpdateConfig updates the payment configuration.
// PUT /api/v1/admin/payment/config
func (h *PaymentHandler) UpdateConfig(c *gin.Context) {
	var req service.UpdatePaymentConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	executeAdminIdempotentJSON(c, "admin.payment.config.update", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		if err := h.configService.UpdatePaymentConfig(ctx, req); err != nil {
			return nil, err
		}
		return gin.H{"message": "updated"}, nil
	})
}
