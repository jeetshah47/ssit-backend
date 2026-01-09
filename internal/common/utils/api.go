package utils

import (
	"net/http"

	"github.com/equitywala/backend/internal/common/errors"
)

// RequestCtx represents a request context interface
type RequestCtx interface {
	GetIP() string
	GetUserID() string
	GetEmail() string
}

// PrepareMsg prepares an error message for logging
func PrepareMsg(err error, bundle interface{}) string {
	if domainErr, ok := err.(*errors.DomainError); ok {
		return domainErr.Error()
	}
	return err.Error()
}

// MapDomainErrorToHTTPStatus maps domain error codes to HTTP status codes
func MapDomainErrorToHTTPStatus(code string) int {
	switch code {
	case "USER_NOT_FOUND", "KYC_NOT_FOUND", "SUBSCRIPTION_NOT_FOUND", "OTP_NOT_FOUND", "SESSION_NOT_FOUND":
		return http.StatusNotFound
	case "INVALID_EMAIL", "INVALID_INPUT", "VALIDATION_ERROR", "INVALID_OTP", "OTP_EXPIRED", "OTP_ALREADY_USED", "OTP_MAX_ATTEMPTS":
		return http.StatusBadRequest
	case "UNAUTHORIZED", "INVALID_TOKEN":
		return http.StatusUnauthorized
	case "FORBIDDEN", "INSUFFICIENT_PERMISSIONS":
		return http.StatusForbidden
	case "USER_ALREADY_EXISTS", "DUPLICATE_ENTRY":
		return http.StatusConflict
	case "PAYMENT_FAILED", "KYC_VERIFICATION_FAILED":
		return http.StatusUnprocessableEntity
	case "DATABASE_ERROR":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

