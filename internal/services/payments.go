package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/equitywala/backend/internal/infrastructure/paytm"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentService handles payment operations
type PaymentService struct {
	paymentRepo         repositories.PaymentRepo
	paytmClient         *paytm.PaytmClient
	selectionRepo       repositories.PaymentPlanSelectionRepo
	packageService      *FindPricingPackageService
	subscriptionService *CreateSubscriptionService
	db                  *gorm.DB
}

// NewPaymentService creates a new payment service
func NewPaymentService(
	paymentRepo repositories.PaymentRepo,
	paytmClient *paytm.PaytmClient,
	selectionRepo repositories.PaymentPlanSelectionRepo,
	packageService *FindPricingPackageService,
	subscriptionService *CreateSubscriptionService,
	db *gorm.DB,
) *PaymentService {
	return &PaymentService{
		paymentRepo:         paymentRepo,
		paytmClient:         paytmClient,
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
// According to Paytm Show Payment Page documentation:
// https://www.paytmpayments.com/docs/show-payment-page
// The frontend needs: mid, orderId, and txnToken to show the payment page
type CreateOrderResponse struct {
	OrderID    string  `json:"order_id"`    // Paytm ORDERID (e.g., "ORDER_xxx")
	TxnToken   string  `json:"txn_token"`   // Transaction token from Initiate Transaction API
	Amount     float64 `json:"amount"`      // Transaction amount
	Currency   string  `json:"currency"`    // Currency code
	MerchantID string  `json:"merchant_id"` // Paytm Merchant ID (MID)
}

// CreateOrder creates a Paytm transaction for the user's selected payment plan
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
			// For existing payments, GatewayOrderID contains the transaction token
			// We need to extract the ORDERID from the payment record
			// The ORDERID format is "ORDER_xxx" which we can reconstruct from payment ID
			paymentIDStr := strings.ReplaceAll(existingPayment.ID.String(), "-", "")
			existingOrderID := fmt.Sprintf("ORDER_%s", paymentIDStr)

			return &CreateOrderResponse{
				OrderID:    existingOrderID,                 // Reconstruct ORDERID from payment ID
				TxnToken:   *existingPayment.GatewayOrderID, // Transaction token stored in GatewayOrderID
				Amount:     existingPayment.Amount,
				Currency:   existingPayment.Currency,
				MerchantID: s.paytmClient.GetMerchantID(),
			}, nil
		}
	}

	// 5. Create payment record with status CREATED
	now := time.Now()
	payment := &models.Payment{
		ID:             uuid.New(),
		UserID:         req.UserID,
		Amount:         packageModel.Price,
		Currency:       packageModel.Currency,
		PaymentMethod:  "paytm", // Using Paytm gateway (user will choose UPI/card/netbanking in checkout)
		PaymentGateway: stringPtr("paytm"),
		Status:         "pending",
		IdempotencyKey: &idempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	payment.MarkAsCreated()

	log.Printf("[PaymentService] Creating payment record - ID: %s, UserID: %s, Amount: %.2f, Currency: %s, PaymentMethod: %s, PaymentGateway: %s, Status: %s",
		payment.ID, payment.UserID, payment.Amount, payment.Currency, payment.PaymentMethod, *payment.PaymentGateway, payment.Status)

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		log.Printf("[PaymentService] ERROR creating payment record - ID: %s, Error: %v", payment.ID, err)
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	log.Printf("[PaymentService] Payment record created successfully - ID: %s", payment.ID)

	// 6. Create Paytm transaction
	// Paytm requires ORDERID - using payment ID as order ID
	uuidStr := strings.ReplaceAll(payment.ID.String(), "-", "")
	orderID := fmt.Sprintf("ORDER_%s", uuidStr)
	// Paytm uses rupees directly, not paise - convert to int64 rupees (multiply by 100 for internal representation)
	amountInRupees := int64(packageModel.Price * 100) // Store as paise internally for consistency

	paytmOrder, err := s.paytmClient.CreateOrder(paytm.CreateOrderRequest{
		Amount:   amountInRupees, // This will be converted to rupees in Paytm client
		Currency: packageModel.Currency,
		Receipt:  orderID,
	})
	if err != nil {
		// Mark payment as failed
		payment.MarkAsFailed(fmt.Sprintf("Failed to create Paytm transaction: %v", err))
		s.paymentRepo.Update(ctx, payment)
		return nil, fmt.Errorf("failed to create Paytm transaction: %w", err)
	}

	// 7. Update payment with gateway order ID (Paytm transaction token)
	payment.GatewayOrderID = &paytmOrder.ID
	payment.MarkAsAttempted()
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment with order ID: %w", err)
	}

	// According to Paytm Show Payment Page documentation:
	// https://www.paytmpayments.com/docs/show-payment-page
	// The frontend needs:
	// - orderId: The ORDERID we sent to Paytm (e.g., "ORDER_xxx")
	// - txnToken: Transaction token from Initiate Transaction API response
	// - mid: Merchant ID
	return &CreateOrderResponse{
		OrderID:    orderID,       // Paytm ORDERID (e.g., "ORDER_xxx")
		TxnToken:   paytmOrder.ID, // Transaction token from Paytm API
		Amount:     packageModel.Price,
		Currency:   packageModel.Currency,
		MerchantID: s.paytmClient.GetMerchantID(), // Paytm Merchant ID (MID)
	}, nil
}

// VerifyPaymentRequest represents the request to verify a payment
type VerifyPaymentRequest struct {
	OrderID   string `json:"order_id"`   // Paytm ORDERID
	PaymentID string `json:"payment_id"` // Paytm TXNID
	Checksum  string `json:"checksum"`   // Paytm CHECKSUMHASH
}

// VerifyPayment verifies the payment signature and creates subscription
func (s *PaymentService) VerifyPayment(ctx context.Context, userID uuid.UUID, req VerifyPaymentRequest) error {
	// 1. Find payment by order ID
	payment, err := s.paymentRepo.FindByOrderID(ctx, req.OrderID)
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
	if !s.paytmClient.VerifyPaymentSignature(
		req.OrderID,
		req.PaymentID,
		req.Checksum,
	) {
		payment.MarkAsFailed("Invalid payment signature")
		s.paymentRepo.Update(ctx, payment)
		return fmt.Errorf("invalid payment signature")
	}

	// 5. Update payment status to VERIFIED
	payment.GatewayPaymentID = &req.PaymentID
	payment.MarkAsVerified()

	// Store payment data
	paymentData := models.PaymentData{
		"paytm_order_id": req.OrderID,   // ORDERID
		"paytm_txn_id":   req.PaymentID, // TXNID
		"paytm_checksum": req.Checksum,  // CHECKSUMHASH
		"verified_at":    time.Now().Format(time.RFC3339),
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

// WebhookEvent represents a Paytm webhook event
type WebhookEvent struct {
	Event   string                 `json:"event"`
	Payload map[string]interface{} `json:"payload"`
}

// HandleWebhook handles Paytm webhook events
func (s *PaymentService) HandleWebhook(ctx context.Context, payload string, signature string) error {
	// 1. Verify webhook signature
	if !s.paytmClient.VerifyWebhookSignature(payload, signature) {
		return fmt.Errorf("invalid webhook signature")
	}

	// 2. Parse webhook payload (Paytm format)
	var webhookData map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &webhookData); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// 3. Extract Paytm callback parameters
	// Paytm sends ORDERID, TXNID, TXNAMOUNT, STATUS, etc.
	var orderID, paymentID, status string
	if id, ok := webhookData["ORDERID"].(string); ok {
		orderID = id
	}
	if id, ok := webhookData["TXNID"].(string); ok {
		paymentID = id
	}
	if statusVal, ok := webhookData["STATUS"].(string); ok {
		status = statusVal
	}

	if orderID == "" {
		return fmt.Errorf("ORDERID not found in webhook payload")
	}

	// 8. Find payment by order ID
	payment, err := s.paymentRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to find payment: %w", err)
	}
	if payment == nil {
		return fmt.Errorf("payment not found for order ID: %s", orderID)
	}

	// 4. Handle Paytm transaction status
	// Paytm STATUS values: "TXN_SUCCESS", "TXN_FAILURE", "PENDING", etc.
	if payment.PaymentData == nil {
		paymentData := make(models.PaymentData)
		payment.PaymentData = &paymentData
	}

	// Store all webhook data
	for k, v := range webhookData {
		(*payment.PaymentData)[k] = v
	}
	(*payment.PaymentData)["webhook_received_at"] = time.Now().Format(time.RFC3339)

	switch status {
	case "TXN_SUCCESS":
		// Payment was successful
		if payment.CanTransitionToConfirmed() {
			payment.MarkAsConfirmed()
			if paymentID != "" {
				payment.GatewayPaymentID = &paymentID
			}

			if err := s.paymentRepo.Update(ctx, payment); err != nil {
				return fmt.Errorf("failed to update payment: %w", err)
			}

			// Ensure subscription is active (idempotent)
			if payment.SubscriptionID != nil {
				// Subscription already created, just ensure it's active
				// This is handled by the subscription service
			}
		}

	case "TXN_FAILURE":
		// Payment failed
		reason := "Payment failed"
		if msg, ok := webhookData["RESPMSG"].(string); ok {
			reason = msg
		}
		payment.MarkAsFailed(reason)
		if err := s.paymentRepo.Update(ctx, payment); err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

	case "PENDING":
		// Payment is pending
		// Keep payment in processing state
		payment.MarkAsAttempted()
		if err := s.paymentRepo.Update(ctx, payment); err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

	default:
		// Unknown status - log but don't fail
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
