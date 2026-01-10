package razorpay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/equitywala/backend/internal/core/config"
	razorpaySDK "github.com/razorpay/razorpay-go"
)

// RazorpayClient wraps the Razorpay SDK client
type RazorpayClient struct {
	client *razorpaySDK.Client
	config config.RazorpayConfig
}

// NewRazorpayClient creates a new Razorpay client
func NewRazorpayClient(cfg config.RazorpayConfig) *RazorpayClient {
	razorpayClient := razorpaySDK.NewClient(cfg.GetKeyID(), cfg.GetKeySecret())
	return &RazorpayClient{
		client: razorpayClient,
		config: cfg,
	}
}

// Order represents a Razorpay order
type Order struct {
	ID     string
	Amount int64
	Status string
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	Amount   int64  // Amount in paise
	Currency string // Currency code (e.g., "INR")
	Receipt  string // Unique receipt ID
}

// CreateOrder creates a new Razorpay order
func (r *RazorpayClient) CreateOrder(req CreateOrderRequest) (*Order, error) {
	data := map[string]interface{}{
		"amount":   req.Amount,
		"currency": req.Currency,
		"receipt":  req.Receipt,
	}

	order, err := r.client.Order.Create(data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Razorpay order: %w", err)
	}

	// Extract order details
	orderID, ok := order["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid order ID in response")
	}

	amount, ok := order["amount"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid amount in response")
	}

	status, _ := order["status"].(string)

	return &Order{
		ID:     orderID,
		Amount: int64(amount),
		Status: status,
	}, nil
}

// VerifyPaymentSignature verifies the payment signature using HMAC SHA256
func (r *RazorpayClient) VerifyPaymentSignature(orderID, paymentID, signature string) bool {
	// Create the message to verify
	message := orderID + "|" + paymentID

	// Calculate expected signature
	expectedSignature := r.calculateHMAC(message, r.config.GetKeySecret())

	// Use constant-time comparison to prevent timing attacks
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// VerifyWebhookSignature verifies the webhook signature
func (r *RazorpayClient) VerifyWebhookSignature(payload, signature string) bool {
	// Calculate expected signature
	expectedSignature := r.calculateHMAC(payload, r.config.GetWebhookSecret())

	// Use constant-time comparison to prevent timing attacks
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// calculateHMAC calculates HMAC SHA256 signature
func (r *RazorpayClient) calculateHMAC(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// GetKeyID returns the Razorpay key ID for frontend
func (r *RazorpayClient) GetKeyID() string {
	return r.config.GetKeyID()
}

// ConvertRupeesToPaise converts rupees to paise
func ConvertRupeesToPaise(rupees float64) int64 {
	return int64(rupees * 100)
}

// ConvertPaiseToRupees converts paise to rupees
func ConvertPaiseToRupees(paise int64) float64 {
	return float64(paise) / 100.0
}

// FormatAmount formats amount as string for Razorpay API
func FormatAmount(amount float64) string {
	paise := ConvertRupeesToPaise(amount)
	return strconv.FormatInt(paise, 10)
}

