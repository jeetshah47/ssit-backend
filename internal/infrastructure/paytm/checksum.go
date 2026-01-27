package paytm

import (
	"strings"

	PaytmChecksum "github.com/paytm/Paytm_Go_Checksum/paytm"
)

// GenerateChecksum generates Paytm checksum using the official Paytm Go checksum library
// This uses Paytm's official checksum utility for JSON-based APIs
// Reference: https://github.com/paytm/Paytm_Go_Checksum
// Reference: https://www.paytmpayments.com/docs/checksum-implementation
//
// For JSON-based APIs (like Initiate Transaction API):
// - Use GenerateSignatureByString with the JSON body string
// - Algorithm: HMAC SHA256
// - Output: Base64 encoded string
func GenerateChecksum(body string, merchantKey string) string {
	// Trim spaces from merchant key (important for correct checksum generation)
	key := strings.TrimSpace(merchantKey)

	if key == "" {
		return ""
	}

	if body == "" {
		return ""
	}

	// Use official Paytm library's GenerateSignatureByString for JSON body
	// This is the recommended method for JSON-based APIs like Initiate Transaction
	return PaytmChecksum.GenerateSignatureByString(body, key)
}

// VerifyChecksum verifies Paytm checksum using the official Paytm Go checksum library
// This uses Paytm's official checksum utility for JSON-based APIs
// Reference: https://github.com/paytm/Paytm_Go_Checksum
// Reference: https://www.paytmpayments.com/docs/checksum-implementation
func VerifyChecksum(message, merchantKey, receivedChecksum string) bool {
	if merchantKey == "" || message == "" || receivedChecksum == "" {
		return false
	}

	// Trim spaces from merchant key
	key := strings.TrimSpace(merchantKey)

	// Use official Paytm library's VerifySignatureByString for JSON body
	// This is the recommended method for JSON-based APIs
	return PaytmChecksum.VerifySignatureByString(message, key, receivedChecksum)
}

// GenerateChecksumFromParams generates checksum from sorted parameters (NVP format)
// This is used for older Paytm APIs that use Name-Value Pair format
// Uses the official Paytm Go checksum library
// Reference: https://github.com/paytm/Paytm_Go_Checksum
// Reference: https://www.paytmpayments.com/docs/checksum-implementation
func GenerateChecksumFromParams(params map[string]string, merchantKey string) string {
	// Trim spaces from merchant key
	key := strings.TrimSpace(merchantKey)

	if key == "" {
		return ""
	}

	// Remove CHECKSUMHASH from params if present (should not be included in checksum generation)
	cleanParams := make(map[string]string)
	for k, v := range params {
		if k != "CHECKSUMHASH" && k != "checksumhash" {
			cleanParams[k] = v
		}
	}

	// Use official Paytm library's GenerateSignature for parameter map (NVP format)
	// This automatically sorts parameters alphabetically and generates checksum
	return PaytmChecksum.GenerateSignature(cleanParams, key)
}
