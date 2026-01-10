package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/equitywala/backend/internal/infrastructure/razorpay"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentService handles payment operations
type PaymentService struct {
	paymentRepo           repositories.PaymentRepo
	razorpayClient        *razorpay.RazorpayClient
	selectionRepo         repositories.PaymentPlanSelectionRepo
	packageService        *FindPricingPackageService
	subscriptionService   *CreateSubscriptionService
	db                    *gorm.DB
}

// NewPaymentService creates a new payment service
func NewPaymentService(
	paymentRepo repositories.PaymentRepo,
	razorpayClient *razorpay.RazorpayClient,
	selectionRepo repositories.PaymentPlanSelectionRepo,
	packageService *FindPricingPackageService,
	subscriptionService *CreateSubscriptionService,
	db *gorm.DB,
) *PaymentService {
	return &PaymentService{
		paymentRepo:         paymentRepo,
		razorpayClient:      razorpayClient,
		selectionRepo:       selectionRepo,
		packageService:      packageService,
		subscriptionService: subscriptionService,
		db:                  db,
	}
}

// CreateOrderRequest represents the request to create a payment order
type CreateOrderRequest struct {
	UserID uuid.UUID
}

// CreateOrderResponse represents the response from creating an order
type CreateOrderResponse struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
	Currency string `json:"currency"`
	KeyID   string  `json:"key_id"`
}

// CreateOrder creates a Razorpay order for the user's selected payment plan
func (s *PaymentService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	// 1. Get user's payment plan selection
	selection, err := s.selectionRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment plan selection: %w", err)
	}
	if selection == nil {
		return nil, fmt.Errorf("no payment plan selected")
	}
	if selection.Status != "pending" {
		return nil, fmt.Errorf("payment plan selection is not pending")
	}

	// 2. Get pricing package details
	var packageModel models.PricingPackageModel
	if err := s.db.WithContext(ctx).
		Where("id = ?", selection.PackageID).
		First(&packageModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("pricing package not found")
		}
		return nil, fmt.Errorf("failed to get pricing package: %w", err)
	}

	// 3. Generate idempotency key
	idempotencyKey := generateIdempotencyKey()

	// 4. Check for existing payment with same idempotency key (idempotency check)
	existingPayment, err := s.paymentRepo.FindByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check idempotency: %w", err)
	}
	if existingPayment != nil {
		// Return existing order if payment already created
		if existingPayment.GatewayOrderID != nil {
			return &CreateOrderResponse{
				OrderID:  *existingPayment.GatewayOrderID,
				Amount:   existingPayment.Amount,
				Currency: existingPayment.Currency,
				KeyID:    s.razorpayClient.GetKeyID(),
			}, nil
		}
	}

	// 5. Create payment record with status CREATED
	now := time.Now()
	payment := &models.Payment{
		ID:            uuid.New(),
		UserID:        req.UserID,
		Amount:        packageModel.Price,
		Currency:      packageModel.Currency,
		PaymentMethod: "razorpay", // Using Razorpay gateway (user will choose UPI/card/netbanking in checkout)
		PaymentGateway: stringPtr("razorpay"),
		Status:        "pending",
		IdempotencyKey: &idempotencyKey,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	payment.MarkAsCreated()

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// 6. Create Razorpay order
	// Razorpay requires receipt ID to be max 40 characters
	// UUID without dashes is 32 characters, which is safe
	uuidStr := strings.ReplaceAll(payment.ID.String(), "-", "")
	receiptID := fmt.Sprintf("rcpt_%s", uuidStr)
	// receiptID is now "rcpt_" (5) + UUID without dashes (32) = 37 characters, which is < 40
	amountInPaise := razorpay.ConvertRupeesToPaise(packageModel.Price)

	razorpayOrder, err := s.razorpayClient.CreateOrder(razorpay.CreateOrderRequest{
		Amount:   amountInPaise,
		Currency: packageModel.Currency,
		Receipt:  receiptID,
	})
	if err != nil {
		// Mark payment as failed
		payment.MarkAsFailed(fmt.Sprintf("Failed to create Razorpay order: %v", err))
		s.paymentRepo.Update(ctx, payment)
		return nil, fmt.Errorf("failed to create Razorpay order: %w", err)
	}

	// 7. Update payment with gateway order ID
	payment.GatewayOrderID = &razorpayOrder.ID
	payment.MarkAsAttempted()
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment with order ID: %w", err)
	}

	return &CreateOrderResponse{
		OrderID:  razorpayOrder.ID,
		Amount:   packageModel.Price,
		Currency: packageModel.Currency,
		KeyID:    s.razorpayClient.GetKeyID(),
	}, nil
}

// VerifyPaymentRequest represents the request to verify a payment
type VerifyPaymentRequest struct {
	RazorpayOrderID   string `json:"razorpay_order_id"`
	RazorpayPaymentID string `json:"razorpay_payment_id"`
	RazorpaySignature string `json:"razorpay_signature"`
}

// VerifyPayment verifies the payment signature and creates subscription
func (s *PaymentService) VerifyPayment(ctx context.Context, userID uuid.UUID, req VerifyPaymentRequest) error {
	// 1. Find payment by order ID
	payment, err := s.paymentRepo.FindByOrderID(ctx, req.RazorpayOrderID)
	if err != nil {
		return fmt.Errorf("failed to find payment: %w", err)
	}
	if payment == nil {
		return fmt.Errorf("payment not found")
	}

	// 2. Verify user owns this payment
	if payment.UserID != userID {
		return fmt.Errorf("unauthorized: payment does not belong to user")
	}

	// 3. Check if already verified/confirmed (idempotency)
	if payment.IsSuccess() {
		// Payment already verified, return success
		return nil
	}

	// 4. Verify signature
	if !s.razorpayClient.VerifyPaymentSignature(
		req.RazorpayOrderID,
		req.RazorpayPaymentID,
		req.RazorpaySignature,
	) {
		payment.MarkAsFailed("Invalid payment signature")
		s.paymentRepo.Update(ctx, payment)
		return fmt.Errorf("invalid payment signature")
	}

	// 5. Update payment status to VERIFIED
	payment.GatewayPaymentID = &req.RazorpayPaymentID
	payment.MarkAsVerified()

	// Store payment data
	paymentData := models.PaymentData{
		"razorpay_order_id":   req.RazorpayOrderID,
		"razorpay_payment_id": req.RazorpayPaymentID,
		"razorpay_signature":  req.RazorpaySignature,
		"verified_at":         time.Now().Format(time.RFC3339),
	}
	payment.PaymentData = &paymentData

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// 6. Create subscription in a transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get payment plan selection
	selection, err := s.selectionRepo.FindByUserID(ctx, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get payment plan selection: %w", err)
	}

	// Get pricing package
	var packageModel models.PricingPackageModel
	if err := tx.Where("id = ?", selection.PackageID).First(&packageModel).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get pricing package: %w", err)
	}

	// Create subscription
	subscriptionCmd := CreateSubscriptionCmd{
		UserID:       userID,
		PackageID:    selection.PackageID,
		PaymentID:    &payment.ID,
		Price:        packageModel.Price,
		Currency:     packageModel.Currency,
		DurationDays: packageModel.DurationDays,
	}

	_, err = s.subscriptionService.ExecuteWithTx(ctx, tx, subscriptionCmd)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	// Mark selection as confirmed
	selection.Confirm()
	if err := s.selectionRepo.Update(ctx, selection); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update selection: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WebhookEvent represents a Razorpay webhook event
type WebhookEvent struct {
	Event   string                 `json:"event"`
	Payload map[string]interface{} `json:"payload"`
}

// HandleWebhook handles Razorpay webhook events
func (s *PaymentService) HandleWebhook(ctx context.Context, payload string, signature string) error {
	// 1. Verify webhook signature
	if !s.razorpayClient.VerifyWebhookSignature(payload, signature) {
		return fmt.Errorf("invalid webhook signature")
	}

	// 2. Parse webhook payload
	var webhookEvent map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &webhookEvent); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// 3. Extract event type
	event, ok := webhookEvent["event"].(string)
	if !ok {
		return fmt.Errorf("invalid event type in webhook payload")
	}

	// 4. Extract payload entity
	payloadEntity, ok := webhookEvent["payload"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payload structure in webhook")
	}

	// 5. Extract payment entity
	paymentEntity, ok := payloadEntity["payment"].(map[string]interface{})
	if !ok {
		// Some events might not have payment entity (e.g., order.paid)
		paymentEntity = make(map[string]interface{})
	}

	// 6. Extract order entity
	orderEntity, ok := payloadEntity["order"].(map[string]interface{})
	if !ok {
		orderEntity = make(map[string]interface{})
	}

	// 7. Extract order ID and payment ID
	var orderID, paymentID string
	if id, ok := orderEntity["id"].(string); ok {
		orderID = id
	}
	// Try to get payment ID from payment entity
	if pid, ok := paymentEntity["id"].(string); ok {
		paymentID = pid
	}
	// Also check payment.entity structure
	if entity, ok := paymentEntity["entity"].(map[string]interface{}); ok {
		if pid, ok := entity["id"].(string); ok {
			paymentID = pid
		}
	}

	if orderID == "" {
		return fmt.Errorf("order ID not found in webhook payload")
	}

	// 8. Find payment by order ID
	payment, err := s.paymentRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to find payment: %w", err)
	}
	if payment == nil {
		return fmt.Errorf("payment not found for order ID: %s", orderID)
	}

	// 9. Handle different event types
	switch event {
	case "payment.captured":
		// Payment was successfully captured
		if payment.CanTransitionToConfirmed() {
			payment.MarkAsConfirmed()
			if paymentID != "" {
				payment.GatewayPaymentID = &paymentID
			}
			
			// Update payment data with webhook info
			if payment.PaymentData == nil {
				paymentData := make(models.PaymentData)
				payment.PaymentData = &paymentData
			}
			(*payment.PaymentData)["webhook_event"] = event
			(*payment.PaymentData)["webhook_received_at"] = time.Now().Format(time.RFC3339)
			
			if err := s.paymentRepo.Update(ctx, payment); err != nil {
				return fmt.Errorf("failed to update payment: %w", err)
			}

			// Ensure subscription is active (idempotent)
			if payment.SubscriptionID != nil {
				// Subscription already created, just ensure it's active
				// This is handled by the subscription service
			}
		}

	case "payment.failed":
		// Payment failed
		reason := "Payment failed"
		if msg, ok := paymentEntity["error_description"].(string); ok {
			reason = msg
		}
		payment.MarkAsFailed(reason)
		if err := s.paymentRepo.Update(ctx, payment); err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

	case "order.paid":
		// Order was paid (alternative to payment.captured)
		if payment.CanTransitionToConfirmed() {
			payment.MarkAsConfirmed()
			if paymentID != "" {
				payment.GatewayPaymentID = &paymentID
			}
			
			if payment.PaymentData == nil {
				paymentData := make(models.PaymentData)
				payment.PaymentData = &paymentData
			}
			(*payment.PaymentData)["webhook_event"] = event
			(*payment.PaymentData)["webhook_received_at"] = time.Now().Format(time.RFC3339)
			
			if err := s.paymentRepo.Update(ctx, payment); err != nil {
				return fmt.Errorf("failed to update payment: %w", err)
			}
		}

	default:
		// Log unknown event types but don't fail
		// In production, log this for monitoring
	}

	return nil
}

// Helper functions

func generateIdempotencyKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func stringPtr(s string) *string {
	return &s
}

