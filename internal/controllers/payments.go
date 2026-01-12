package controllers

import (
	"fmt"

	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/services"
	"github.com/google/uuid"
)

// PaymentController handles payment endpoints
type PaymentController struct {
	paymentService *services.PaymentService
}

// NewPaymentController creates a new payment controller
func NewPaymentController(paymentService *services.PaymentService) *PaymentController {
	return &PaymentController{
		paymentService: paymentService,
	}
}

// CreateOrder handles payment order creation
func (c *PaymentController) CreateOrder(ctx *utils.Context) (interface{}, error) {
	// Get user ID from context
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Create order
	req := services.CreateOrderRequest{
		UserID: userID,
	}

	fmt.Printf("[PaymentController] Creating order for user: %s\n", userID)
	response, err := c.paymentService.CreateOrder(ctx.Request.Context(), req)
	if err != nil {
		fmt.Printf("[PaymentController] ERROR creating order for user %s: %v\n", userID, err)
		return nil, err
	}

	fmt.Printf("[PaymentController] Order created successfully - OrderID: %s, Amount: %.2f\n", response.OrderID, response.Amount)
	return response, nil
}

// VerifyPayment handles payment verification
func (c *PaymentController) VerifyPayment(ctx *utils.Context) (interface{}, error) {
	// Get user ID from context
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Parse request
	var req services.VerifyPaymentRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Validate request
	if req.OrderID == "" {
		return nil, fmt.Errorf("order_id is required")
	}
	if req.PaymentID == "" {
		return nil, fmt.Errorf("payment_id is required")
	}
	if req.Checksum == "" {
		return nil, fmt.Errorf("checksum is required")
	}

	// Verify payment
	if err := c.paymentService.VerifyPayment(ctx.Request.Context(), userID, req); err != nil {
		return nil, err
	}

	return models.VerifyPaymentResponse{
		Message: "Payment verified successfully",
		Success: true,
	}, nil
}

// HandleWebhook handles Paytm webhook events
// NOTE: Webhook feature is currently commented out - will be enabled in production
/*
func (c *PaymentController) HandleWebhook(ctx *utils.Context) (interface{}, error) {
	// Get webhook signature from header (Paytm uses different header)
	signature := ctx.GetHeader("X-Paytm-Signature")
	if signature == "" {
		// Try alternative header name
		signature = ctx.GetHeader("CHECKSUMHASH")
	}
	if signature == "" {
		return nil, fmt.Errorf("missing webhook signature")
	}

	// Read raw body
	bodyBytes, err := ctx.GetRawData()
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	payload := string(bodyBytes)

	// Handle webhook asynchronously (non-blocking)
	// In production, use a proper job queue
	go func() {
		if err := c.paymentService.HandleWebhook(ctx.Request.Context(), payload, signature); err != nil {
			// Log error - in production, use proper logging
			fmt.Printf("Webhook processing error: %v\n", err)
		}
	}()

	// Return 200 immediately
	return models.WebhookResponse{
		Message: "Webhook received",
		Success: true,
	}, nil
}
*/

