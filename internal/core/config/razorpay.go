package config

// RazorpayConfig holds Razorpay configuration
type RazorpayConfig struct {
	KeyID        string
	KeySecret    string
	WebhookSecret string
}

// GetKeyID returns the Razorpay key ID
func (r RazorpayConfig) GetKeyID() string {
	return r.KeyID
}

// GetKeySecret returns the Razorpay key secret
func (r RazorpayConfig) GetKeySecret() string {
	return r.KeySecret
}

// GetWebhookSecret returns the Razorpay webhook secret
func (r RazorpayConfig) GetWebhookSecret() string {
	return r.WebhookSecret
}

// newRazorpayConfig creates a new Razorpay configuration from environment variables
func newRazorpayConfig() RazorpayConfig {
	return RazorpayConfig{
		KeyID:        getEnv("RAZORPAY_KEY_ID", ""),
		KeySecret:    getEnv("RAZORPAY_KEY_SECRET", ""),
		WebhookSecret: getEnv("RAZORPAY_WEBHOOK_SECRET", ""),
	}
}

