package repositories

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentRepo defines the interface for payment data persistence
type PaymentRepo interface {
	// Create creates a new payment
	Create(ctx context.Context, payment *models.Payment) error

	// FindByID finds a payment by ID
	FindByID(ctx context.Context, id uuid.UUID) (*models.Payment, error)

	// FindByOrderID finds a payment by Paytm order ID
	FindByOrderID(ctx context.Context, orderID string) (*models.Payment, error)

	// FindByUserID finds all payments for a user
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Payment, error)

	// Update updates an existing payment
	Update(ctx context.Context, payment *models.Payment) error

	// FindByIdempotencyKey finds a payment by idempotency key
	FindByIdempotencyKey(ctx context.Context, key string) (*models.Payment, error)
}

// paymentRepo implements the payment repository interface
type paymentRepo struct {
	*RepoContext
}

// NewPaymentRepo creates a new payment repository
func NewPaymentRepo(ctx *RepoContext) PaymentRepo {
	return &paymentRepo{RepoContext: ctx}
}

// Create creates a new payment
func (r *paymentRepo) Create(ctx context.Context, payment *models.Payment) error {
	log.Printf("[PaymentRepo] Attempting to create payment - ID: %s, UserID: %s, PaymentMethod: %s, PaymentGateway: %v, Amount: %.2f, Currency: %s, Status: %s",
		payment.ID, payment.UserID, payment.PaymentMethod, payment.PaymentGateway, payment.Amount, payment.Currency, payment.Status)

	if err := r.db.WithContext(ctx).Create(payment).Error; err != nil {
		log.Printf("[PaymentRepo] ERROR creating payment - ID: %s, Error: %v, Error Type: %T", payment.ID, err, err)
		
		// Check if it's a constraint violation
		if strings.Contains(err.Error(), "payments_method_check") {
			log.Printf("[PaymentRepo] CONSTRAINT VIOLATION - payment_method value '%s' is not allowed by constraint", payment.PaymentMethod)
			
			// Try to get the actual constraint definition
			var constraintDef string
			if queryErr := r.db.WithContext(ctx).Raw(`
				SELECT pg_get_constraintdef(oid) 
				FROM pg_constraint 
				WHERE conrelid = 'payments'::regclass 
				AND conname = 'payments_method_check'
			`).Scan(&constraintDef).Error; queryErr == nil {
				log.Printf("[PaymentRepo] Current constraint definition: %s", constraintDef)
			}
		}
		
		return fmt.Errorf("failed to create payment: %w", err)
	}
	
	log.Printf("[PaymentRepo] Payment created successfully - ID: %s", payment.ID)
	return nil
}

// FindByID finds a payment by ID
func (r *paymentRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	return &payment, nil
}

// FindByOrderID finds a payment by Razorpay order ID
func (r *paymentRepo) FindByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.WithContext(ctx).
		Where("gateway_order_id = ?", orderID).
		First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find payment by order ID: %w", err)
	}
	return &payment, nil
}

// FindByUserID finds all payments for a user
func (r *paymentRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Payment, error) {
	var payments []*models.Payment
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to find payments by user ID: %w", err)
	}
	return payments, nil
}

// Update updates an existing payment
func (r *paymentRepo) Update(ctx context.Context, payment *models.Payment) error {
	if err := r.db.WithContext(ctx).Save(payment).Error; err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}
	return nil
}

// FindByIdempotencyKey finds a payment by idempotency key
func (r *paymentRepo) FindByIdempotencyKey(ctx context.Context, key string) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.WithContext(ctx).
		Where("idempotency_key = ?", key).
		First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find payment by idempotency key: %w", err)
	}
	return &payment, nil
}

