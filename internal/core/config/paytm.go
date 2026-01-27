package config

import "strings"

// PaytmConfig holds Paytm configuration
type PaytmConfig struct {
	MerchantID    string // Paytm Merchant ID (MID)
	MerchantKey   string // Paytm Merchant Key
	WebhookSecret string // Paytm Webhook Secret
	Environment   string // Paytm environment: "staging" or "production"
	WebsiteName   string // Paytm website name
	CallbackURL   string // Paytm callback URL
}

// GetMerchantID returns the Paytm Merchant ID (MID)
func (p PaytmConfig) GetMerchantID() string {
	return p.MerchantID
}

// GetMerchantKey returns the Paytm Merchant Key
func (p PaytmConfig) GetMerchantKey() string {
	return p.MerchantKey
}

// GetWebhookSecret returns the Paytm webhook secret
func (p PaytmConfig) GetWebhookSecret() string {
	return p.WebhookSecret
}

// GetEnvironment returns the Paytm environment
func (p PaytmConfig) GetEnvironment() string {
	return p.Environment
}

// GetWebsiteName returns the Paytm website name
func (p PaytmConfig) GetWebsiteName() string {
	if p.WebsiteName == "" {
		return "DEFAULT"
	}
	return p.WebsiteName
}

// GetCallbackURL returns the Paytm callback URL
func (p PaytmConfig) GetCallbackURL() string {
	return p.CallbackURL
}

// newPaytmConfig creates a new Paytm configuration from environment variables
func newPaytmConfig() PaytmConfig {
	// Trim spaces from all environment variables when reading
	// This ensures no accidental whitespace causes issues, especially for merchant key
	return PaytmConfig{
		MerchantID:    strings.TrimSpace(getEnv("PAYTM_MERCHANT_ID", "")),
		MerchantKey:   strings.TrimSpace(getEnv("PAYTM_MERCHANT_KEY", "")),
		WebhookSecret: strings.TrimSpace(getEnv("PAYTM_WEBHOOK_SECRET", "")),
		Environment:   strings.TrimSpace(getEnv("PAYTM_ENVIRONMENT", "staging")),
		WebsiteName:   strings.TrimSpace(getEnv("PAYTM_WEBSITE_NAME", "DEFAULT")),
		CallbackURL:   strings.TrimSpace(getEnv("PAYTM_CALLBACK_URL", "")),
	}
}
