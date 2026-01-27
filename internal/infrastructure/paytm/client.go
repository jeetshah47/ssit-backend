package paytm

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/equitywala/backend/internal/core/config"
	PaytmChecksum "github.com/paytm/Paytm_Go_Checksum/paytm"
)

// PaytmClient wraps the Paytm API client
type PaytmClient struct {
	httpClient *http.Client
	config     config.PaytmConfig
	baseURL    string // Paytm API base URL (staging or production)
}

// NewPaytmClient creates a new Paytm client
func NewPaytmClient(cfg config.PaytmConfig) *PaytmClient {
	// Determine base URL based on environment
	// Default to staging/testing environment for safety
	// Note: Using securegw-stage.paytm.in for staging (theia API endpoint)
	// The securestage.paytmpayments.com domain may be for different APIs
	env := cfg.GetEnvironment()
	baseURL := "https://securegw-stage.paytm.in" // Staging/testing environment

	if env == "production" {
		baseURL = "https://securegw.paytm.in" // Production environment
		// log.Printf("[PaytmClient] WARNING: Using PRODUCTION Paytm environment")
	} else {
		// Force staging/testing environment
		baseURL = "https://securegw-stage.paytm.in"
		// log.Printf("[PaytmClient] Using STAGING/TESTING Paytm environment (baseURL: %s)", baseURL)
	}

	return &PaytmClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config:  cfg,
		baseURL: baseURL,
	}
}

// getEnv is a helper to get environment variable
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Order represents a Paytm order/transaction
type Order struct {
	ID     string // Paytm transaction token (TXN_TOKEN)
	Amount int64  // Amount in rupees (not paise for Paytm)
	Status string
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	Amount   int64  // Amount in rupees (Paytm uses rupees, not paise)
	Currency string // Currency code (e.g., "INR")
	Receipt  string // Unique receipt ID (will be used as ORDERID)
}

// PaytmInitiateTransactionRequest represents Paytm's initiate transaction request
// Structure matches Paytm API documentation: https://www.paytmpayments.com/docs/api/initiate-transaction-api
// Field order is important for checksum generation - must match docs example order
type PaytmInitiateTransactionRequest struct {
	Body struct {
		RequestType string `json:"requestType"` // Required: "Payment"
		Mid         string `json:"mid"`         // Required: Merchant ID
		WebsiteName string `json:"websiteName"` // Required: Website name from Paytm dashboard
		OrderID     string `json:"orderId"`     // Required: Unique order identifier
		TxnAmount   struct {
			Value    string `json:"value"`    // Required: Amount as string with 2 decimals (e.g., "1.00")
			Currency string `json:"currency"` // Required: Currency code (e.g., "INR")
		} `json:"txnAmount"`
		UserInfo struct {
			CustID string `json:"custId"` // Required: Customer identifier
		} `json:"userInfo"`
		CallbackURL string `json:"callbackUrl"` // Required: Callback URL for payment response
	} `json:"body"`
	Head struct {
		// According to Paytm docs, only signature is required in head for request
		// Other fields (clientId, version, channelId, requestTimestamp) are optional
		// and may be included in response but not required in request
		Signature string `json:"signature"` // Required: HMAC SHA256 checksum of body JSON
	} `json:"head"`
}

// PaytmInitiateTransactionResponse represents Paytm's initiate transaction response
type PaytmInitiateTransactionResponse struct {
	Head struct {
		ResponseTimestamp string `json:"responseTimestamp"`
		Version           string `json:"version"`
		ClientID          string `json:"clientId"`
		Signature         string `json:"signature"`
	} `json:"head"`
	Body struct {
		ResultInfo struct {
			ResultStatus string `json:"resultStatus"`
			ResultCode   string `json:"resultCode"`
			ResultMsg    string `json:"resultMsg"`
		} `json:"resultInfo"`
		TxnToken         string `json:"txnToken"`
		IsPromoCodeValid bool   `json:"isPromoCodeValid"`
		Authenticated    bool   `json:"authenticated"`
	} `json:"body"`
}

// CreateOrder creates a new Paytm transaction (initiate transaction)
func (p *PaytmClient) CreateOrder(req CreateOrderRequest) (*Order, error) {
	// Paytm requires ORDERID, MID, amount, and other details
	orderID := req.Receipt
	mid := p.config.GetMerchantID()
	merchantKey := p.config.GetMerchantKey()

	// Validate required fields
	if mid == "" {
		// log.Printf("[PaytmClient] ERROR: Paytm Merchant ID (MID) is empty. Please set PAYTM_MERCHANT_ID environment variable")
		return nil, fmt.Errorf("Paytm Merchant ID (MID) is not configured. Please set PAYTM_MERCHANT_ID environment variable")
	}
	if merchantKey == "" {
		// log.Printf("[PaytmClient] ERROR: Paytm Merchant Key is empty. Please set PAYTM_MERCHANT_KEY environment variable")
		return nil, fmt.Errorf("Paytm Merchant Key is not configured. Please set PAYTM_MERCHANT_KEY environment variable")
	}

	// Log MID and key info (masked for security)
	// midMasked := maskString(mid)
	// keyMasked := maskString(merchantKey)
	// log.Printf("[PaytmClient] Creating order - OrderID: %s, MID: %s (length: %d), Key: %s (length: %d), Amount: %d, Currency: %s",
	// 	orderID, midMasked, len(mid), keyMasked, len(merchantKey), req.Amount, req.Currency)

	// Warn if key contains special characters that might need quoting in .env
	// if strings.Contains(merchantKey, "#") || strings.Contains(merchantKey, "%") {
	// 	log.Printf("[PaytmClient] WARNING: Merchant key contains special characters (# or %%). Ensure it's properly quoted in .env file.")
	// }

	// CRITICAL FIX: Use struct to ensure exact field order (Go maps don't preserve order)
	// Paytm requires checksum to be generated from the EXACT JSON bytes that will be sent
	// Field order must match Paytm's expected order: requestType, mid, websiteName, orderId, txnAmount, userInfo, callbackUrl
	callbackURL := p.config.GetCallbackURL()
	if callbackURL == "" {
		callbackURL = getEnv("PAYTM_CALLBACK_URL", "https://yourdomain.com/api/v1/payments/callback")
	}

	// Build body struct with exact field order (matching Paytm's expected order from Postman example)
	// Order: requestType, mid, websiteName, orderId, txnAmount, userInfo, callbackUrl
	type PaytmBodyStruct struct {
		RequestType string `json:"requestType"` // 1st field
		Mid         string `json:"mid"`         // 2nd field
		WebsiteName string `json:"websiteName"` // 3rd field
		OrderID     string `json:"orderId"`     // 4th field
		TxnAmount   struct {
			Value    string `json:"value"` // value before currency (as per Postman)
			Currency string `json:"currency"`
		} `json:"txnAmount"` // 5th field
		UserInfo struct {
			CustID string `json:"custId"`
		} `json:"userInfo"` // 6th field
		CallbackURL string `json:"callbackUrl"` // 7th field
	}

	bodyStruct := PaytmBodyStruct{
		RequestType: "Payment",
		Mid:         mid,
		WebsiteName: p.config.GetWebsiteName(),
		OrderID:     orderID,
	}
	bodyStruct.TxnAmount.Value = fmt.Sprintf("%.2f", float64(req.Amount)/100.0) // Convert paise to rupees
	bodyStruct.TxnAmount.Currency = req.Currency
	bodyStruct.UserInfo.CustID = "CUST_" + orderID
	bodyStruct.CallbackURL = callbackURL

	// Marshal body ONCE - struct preserves field order as defined
	// This is the exact JSON that will be used for checksum AND sent to Paytm
	bodyJSON, err := json.Marshal(bodyStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// This is the exact JSON string that will be used for checksum generation
	bodyJSONStr := string(bodyJSON)

	// Log merchant key info (safely)
	// keyFirstChar := "?"
	// keyLastChar := "?"
	// trimmedKey := strings.TrimSpace(merchantKey)
	// if len(trimmedKey) > 0 {
	// 	keyFirstChar = string(trimmedKey[0])
	// 	keyLastChar = string(trimmedKey[len(trimmedKey)-1])
	// }
	// log.Printf("[PaytmClient] Merchant key length: %d (original: %d), first char: %s, last char: %s",
	// 	len(trimmedKey), len(merchantKey), keyFirstChar, keyLastChar)

	// Generate signature/checksum using HMAC SHA256
	// CRITICAL: Generate checksum from the EXACT bodyJSONStr that will be sent
	// The merchant key is automatically trimmed in GenerateChecksum
	checksum := GenerateChecksum(bodyJSONStr, merchantKey)

	if len(checksum) == 0 {
		return nil, fmt.Errorf("checksum generation failed: empty checksum returned")
	}

	// Log checksum
	log.Printf("[PaytmClient] Generated Checksum: %s", checksum)

	// Log checksum generation details for debugging
	// log.Printf("[PaytmClient] Checksum generation details:")
	// log.Printf("[PaytmClient]   Input JSON length: %d bytes", len(bodyJSONStr))
	// log.Printf("[PaytmClient]   Merchant key length: %d bytes (after trim)", len(strings.TrimSpace(merchantKey)))
	// log.Printf("[PaytmClient]   Algorithm: HMAC SHA256")
	// log.Printf("[PaytmClient]   Output format: Base64 encoded")
	// log.Printf("[PaytmClient]   Generated checksum: %s (length: %d)", checksum, len(checksum))

	// Build final request using struct to ensure "body" comes before "head" (matching Postman example)
	// Using json.RawMessage ensures the body bytes are preserved exactly (no re-marshaling)
	type PaytmRequestStruct struct {
		Body json.RawMessage `json:"body"` // body must come first (as per Postman example)
		Head struct {
			Signature string `json:"signature"`
		} `json:"head"`
	}

	finalRequestStruct := PaytmRequestStruct{
		Body: json.RawMessage(bodyJSON), // Use the EXACT same JSON bytes - no re-marshaling
	}
	finalRequestStruct.Head.Signature = checksum

	// Marshal final request
	// Struct preserves field order: body first, then head
	// json.RawMessage preserves the exact bytes of bodyJSON
	requestJSON, err := json.Marshal(finalRequestStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Log the body JSON used for checksum (before wrapping in final request)
	// This is the exact JSON that checksum was generated from
	wrappedBodyJSON := fmt.Sprintf(`{"body":%s}`, bodyJSONStr)
	log.Printf("[PaytmClient] Request Payload (Body JSON): %s", wrappedBodyJSON)

	// CRITICAL VERIFICATION: Ensure body JSON in final request matches exactly
	// Parse the request JSON using the same struct to verify body is preserved
	var verifyRequest PaytmRequestStruct
	if err := json.Unmarshal(requestJSON, &verifyRequest); err == nil {
		bodyInRequest := string(verifyRequest.Body)
		if bodyInRequest != bodyJSONStr {
			// This should never happen with json.RawMessage, but verify anyway
			return nil, fmt.Errorf("CRITICAL: Body JSON was modified during request construction")
		}
		// Body matches - verification successful
	} else {
		// If unmarshaling fails, log but don't fail (might be a parsing issue)
		// The actual request should still be correct since we used json.RawMessage
	}

	// Log the final payload JSON sent to server
	log.Printf("[PaytmClient] Final Payload JSON: %s", string(requestJSON))

	// CRITICAL: The body JSON in the final request is byte-identical to what we used for checksum
	// because we use json.RawMessage to preserve the exact same JSON bytes
	// log.Printf("[PaytmClient] Body JSON verified: byte-identical match between checksum input and request body")

	// Make HTTP request to Paytm
	// Paytm requires both MID and OrderID as query parameters in the URL
	url := fmt.Sprintf("%s/theia/api/v1/initiateTransaction?mid=%s&orderId=%s", p.baseURL, mid, orderID)
	// log.Printf("[PaytmClient] Making request to Paytm API - Environment: %s, BaseURL: %s, Full URL: %s",
	// 	p.config.GetEnvironment(), p.baseURL, url)
	// log.Printf("[PaytmClient] Request body JSON (length: %d)", len(requestJSON))

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		// log.Printf("[PaytmClient] ERROR calling Paytm API: %v", err)
		return nil, fmt.Errorf("failed to call Paytm API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// log.Printf("[PaytmClient] Paytm API response - Status: %d, Body: %s", resp.StatusCode, string(body))

	// Parse response even if status code is not 200, as Paytm may return error details in JSON
	var paytmResp PaytmInitiateTransactionResponse
	if err := json.Unmarshal(body, &paytmResp); err != nil {
		// If we can't parse the response, return HTTP status error
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Paytm API returned status %d and unparseable response: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("failed to parse Paytm response: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		// Try to extract error message from response
		if paytmResp.Body.ResultInfo.ResultCode != "" {
			return nil, fmt.Errorf("Paytm API returned status %d: %s - %s",
				resp.StatusCode,
				paytmResp.Body.ResultInfo.ResultCode,
				paytmResp.Body.ResultInfo.ResultMsg)
		}
		return nil, fmt.Errorf("Paytm API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Check result status in response body
	if paytmResp.Body.ResultInfo.ResultStatus != "S" {
		// 501 error code typically means system error or endpoint issue
		// if paytmResp.Body.ResultInfo.ResultCode == "501" || paytmResp.Body.ResultInfo.ResultCode == "00000900" {
		// 	log.Printf("[PaytmClient] ERROR: Paytm system error (501). This may indicate:")
		// 	log.Printf("[PaytmClient]   - Incorrect API endpoint URL")
		// 	log.Printf("[PaytmClient]   - Paytm service temporarily unavailable")
		// 	log.Printf("[PaytmClient]   - Request format issue")
		// 	log.Printf("[PaytmClient]   - Merchant credentials issue")
		// }
		return nil, fmt.Errorf("Paytm transaction initiation failed: %s - %s",
			paytmResp.Body.ResultInfo.ResultCode,
			paytmResp.Body.ResultInfo.ResultMsg)
	}

	// Verify response checksum
	// For JSON APIs, Paytm generates checksum from the body JSON only (not the full response)
	// Extract body JSON from response for verification
	// As per Paytm documentation: VerifySignatureByString(body, merchantKey, checksum)
	responseBodyJSON, err := json.Marshal(paytmResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response body for checksum verification: %w", err)
	}
	responseBodyJSONStr := string(responseBodyJSON)

	// Verify checksum using the body JSON string (as per Paytm documentation)
	// Pattern: PaytmChecksum.VerifySignatureByString(body, merchantKey, checksum)
	if !p.verifyPaytmChecksum(responseBodyJSONStr, paytmResp.Head.Signature, merchantKey) {
		return nil, fmt.Errorf("invalid checksum in Paytm response")
	}

	return &Order{
		ID:     paytmResp.Body.TxnToken, // Paytm returns transaction token
		Amount: req.Amount,
		Status: "created",
	}, nil
}

// VerifyPaymentSignature verifies the Paytm payment checksum from callback parameters
// For callback/webhook, Paytm uses parameter map format (NVP - Name-Value Pair)
// Reference: https://www.paytmpayments.com/docs/checksum-implementation
// This follows Paytm's official documentation pattern for parameter-based verification
//
// Pattern from Paytm docs:
// paytmParams := map[string]string{"MID": "...", "ORDER_ID": "...", ...}
// isVerifySignature := PaytmChecksum.VerifySignature(paytmParams, "YOUR_MERCHANT_KEY", paytmChecksum)
func (p *PaytmClient) VerifyPaymentSignature(orderID, paymentID, signature string) bool {
	merchantKey := p.config.GetMerchantKey()

	if merchantKey == "" || orderID == "" || signature == "" {
		return false
	}

	// Build parameter map from callback parameters
	// As per Paytm documentation, all callback parameters should be included
	// Paytm callback typically includes: ORDERID, TXNID, TXNAMOUNT, STATUS, etc.
	params := map[string]string{
		"ORDERID": orderID,
	}
	if paymentID != "" {
		params["TXNID"] = paymentID
	}

	// Use official Paytm library's VerifySignature for parameter map (NVP format)
	// This is the recommended method for callback/webhook verification
	// Pattern: PaytmChecksum.VerifySignature(paytmParams, "YOUR_MERCHANT_KEY", paytmChecksum)
	return PaytmChecksum.VerifySignature(params, merchantKey, signature)
}

// VerifyWebhookSignature verifies the Paytm webhook signature
// For webhooks, Paytm typically uses JSON body format
// Reference: https://www.paytmpayments.com/docs/checksum-implementation
func (p *PaytmClient) VerifyWebhookSignature(payload, signature string) bool {
	merchantKey := p.config.GetWebhookSecret()

	if merchantKey == "" || payload == "" || signature == "" {
		return false
	}

	// Use official Paytm library's VerifySignatureByString for JSON body
	// This is the recommended method for webhook verification with JSON payload
	return PaytmChecksum.VerifySignatureByString(payload, merchantKey, signature)
}

// calculateHMAC calculates HMAC SHA256 signature
func (p *PaytmClient) calculateHMAC(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// Note: Checksum generation is now handled by the standalone utility functions
// in checksum.go (similar to Paytm's official libraries)
// See: GenerateChecksum, VerifyChecksum, GenerateChecksumFromParams

// verifyPaytmChecksum verifies Paytm checksum
// For Paytm, checksum is generated from sorted parameters
func (p *PaytmClient) verifyPaytmChecksum(payload, receivedChecksum, merchantKey string) bool {
	// Use the standalone checksum verification utility
	return VerifyChecksum(payload, merchantKey, receivedChecksum)
}

// generatePaytmChecksumFromParams generates checksum from sorted parameters (Paytm format)
// This is a wrapper around the standalone utility function
func (p *PaytmClient) generatePaytmChecksumFromParams(params map[string]string, merchantKey string) string {
	return GenerateChecksumFromParams(params, merchantKey)
}

// GetMerchantID returns the Paytm Merchant ID for frontend
func (p *PaytmClient) GetMerchantID() string {
	return p.config.GetMerchantID()
}

// ConvertRupeesToPaise converts rupees to paise
func ConvertRupeesToPaise(rupees float64) int64 {
	return int64(rupees * 100)
}

// ConvertPaiseToRupees converts paise to rupees
func ConvertPaiseToRupees(paise int64) float64 {
	return float64(paise) / 100.0
}

// FormatAmount formats amount as string for Paytm API (in rupees)
func FormatAmount(amount float64) string {
	// Paytm uses rupees, not paise
	return fmt.Sprintf("%.2f", amount)
}

// maskString masks a string for logging (shows first 4 and last 4 characters)
func maskString(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
