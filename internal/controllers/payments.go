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

	response, err := c.paymentService.CreateOrder(ctx.Request.Context(), req)
	if err != nil {
		return nil, err
	}

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
	if req.RazorpayOrderID == "" {
		return nil, fmt.Errorf("razorpay_order_id is required")
	}
	if req.RazorpayPaymentID == "" {
		return nil, fmt.Errorf("razorpay_payment_id is required")
	}
	if req.RazorpaySignature == "" {
		return nil, fmt.Errorf("razorpay_signature is required")
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

// HandleWebhook handles Razorpay webhook events
// NOTE: Webhook feature is currently commented out - will be enabled in production
/*
func (c *PaymentController) HandleWebhook(ctx *utils.Context) (interface{}, error) {
	// Get webhook signature from header
	signature := ctx.GetHeader("X-Razorpay-Signature")
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

