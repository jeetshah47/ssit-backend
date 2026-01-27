package utils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/common/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
	case "UNAUTHORIZED", "INVALID_TOKEN", "TOKEN_EXPIRED", "TOKEN_NOT_VALID_YET":
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

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorInfo represents error details in API response
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// HandlerFunc is a function that returns data or error
type HandlerFunc func(ctx *Context) (interface{}, error)

// Context wraps gin.Context with additional utilities
type Context struct {
	*gin.Context
	UserID string // Extracted from auth middleware
	Role   string // Extracted from auth middleware
}

// NewContext creates a new API context from gin context
func NewContext(c *gin.Context) *Context {
	ctx := &Context{Context: c}

	// Extract user ID and role from context (set by middleware)
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			ctx.UserID = id
		}
	}

	if role, exists := c.Get("role"); exists {
		if r, ok := role.(string); ok {
			ctx.Role = r
		}
	}

	return ctx
}

// Handle wraps a handler function and automatically handles response formatting
func Handle(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := NewContext(c)

		// Execute handler
		data, err := handler(ctx)

		// Handle error response
		if err != nil {
			handleError(c, err)
			return
		}

		// Handle success response
		handleSuccess(c, data)
	}
}

// HandleWithStatus wraps a handler function with custom status code
func HandleWithStatus(statusCode int, handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := NewContext(c)

		data, err := handler(ctx)

		if err != nil {
			handleError(c, err)
			return
		}

		handleSuccessWithStatus(c, statusCode, data)
	}
}

// handleSuccess sends a successful response
func handleSuccess(c *gin.Context, data interface{}) {
	response := Response{
		Success: true,
		Data:    data,
	}
	c.JSON(http.StatusOK, response)
}

// handleSuccessWithStatus sends a successful response with custom status code
func handleSuccessWithStatus(c *gin.Context, statusCode int, data interface{}) {
	response := Response{
		Success: true,
		Data:    data,
	}
	c.JSON(statusCode, response)
}

// handleError processes errors and sends appropriate error response
func handleError(c *gin.Context, err error) {
	var statusCode int
	var errorInfo *ErrorInfo

	// Get logger from context if available
	var appLogger logger.Logger
	if loggerVal, exists := c.Get("logger"); exists {
		if l, ok := loggerVal.(logger.Logger); ok {
			appLogger = l
		}
	}

	// Check if it's a validation error (should be 400, not 500)
	if strings.Contains(err.Error(), "validation failed") || strings.Contains(err.Error(), "failed to parse request body") {
		statusCode = http.StatusBadRequest
		errorInfo = &ErrorInfo{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		}
		
		// Log validation errors at warn level
		if appLogger != nil {
			logFields := []interface{}{
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", statusCode,
				"error", err.Error(),
			}
			
			if userID := c.GetString("user_id"); userID != "" {
				logFields = append(logFields, "user_id", userID)
			}
			
			appLogger.Warn("Validation error in API handler", logFields...)
		}
	} else if domainErr, ok := err.(*errors.DomainError); ok {
		statusCode = MapDomainErrorToHTTPStatus(domainErr.Code)
		errorInfo = &ErrorInfo{
			Code:    domainErr.Code,
			Message: domainErr.Message,
		}
		
		// Log domain errors at appropriate levels
		if appLogger != nil {
			logFields := []interface{}{
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", statusCode,
				"error_code", domainErr.Code,
				"error_message", domainErr.Message,
				"error", err.Error(),
			}
			
			// Add user context if available
			if userID := c.GetString("user_id"); userID != "" {
				logFields = append(logFields, "user_id", userID)
			}
			
			// Log at different levels based on status code
			if statusCode >= http.StatusInternalServerError {
				appLogger.Error("Domain error in API handler (5xx)", logFields...)
			} else if statusCode >= http.StatusBadRequest {
				appLogger.Warn("Domain error in API handler (4xx)", logFields...)
			}
		}
	} else {
		// Generic error - always log these as errors
		statusCode = http.StatusInternalServerError
		errorInfo = &ErrorInfo{
			Code:    "INTERNAL_ERROR",
			Message: "An internal error occurred",
		}
		
		// Log all internal server errors with full details
		if appLogger != nil {
			logFields := []interface{}{
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", statusCode,
				"error", err.Error(),
			}
			
			// Add user context if available
			if userID := c.GetString("user_id"); userID != "" {
				logFields = append(logFields, "user_id", userID)
			}
			
			appLogger.Error("Internal server error in API handler", logFields...)
		} else {
			// Fallback to stderr if logger not available
			c.Error(err)
		}
	}

	response := Response{
		Success: false,
		Error:   errorInfo,
	}

	c.JSON(statusCode, response)
}

// BindJSON binds JSON request body to struct and validates
func (c *Context) BindJSON(obj interface{}) error {
	if err := c.Context.ShouldBindJSON(obj); err != nil {
		// Provide more helpful error messages for common validation errors
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			var errorMsgs []string
			for _, fieldErr := range validationErr {
				fieldName := fieldErr.Field()
				tag := fieldErr.Tag()
				
				var msg string
				switch tag {
				case "required":
					msg = fmt.Sprintf("field '%s' is required", fieldName)
				case "email":
					msg = fmt.Sprintf("field '%s' must be a valid email address", fieldName)
				case "uuid":
					msg = fmt.Sprintf("field '%s' must be a valid UUID", fieldName)
				case "min":
					msg = fmt.Sprintf("field '%s' does not meet minimum value requirement", fieldName)
				case "max":
					msg = fmt.Sprintf("field '%s' exceeds maximum value requirement", fieldName)
				default:
					msg = fmt.Sprintf("field '%s' failed validation: %s", fieldName, tag)
				}
				errorMsgs = append(errorMsgs, msg)
			}
			return fmt.Errorf("validation failed: %s", strings.Join(errorMsgs, "; "))
		}
		return fmt.Errorf("failed to parse request body: %w", err)
	}

	// Validate struct tags
	validate := validator.New()
	if err := validate.Struct(obj); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			var errorMsgs []string
			for _, fieldErr := range validationErr {
				fieldName := fieldErr.Field()
				tag := fieldErr.Tag()
				
				var msg string
				switch tag {
				case "required":
					msg = fmt.Sprintf("field '%s' is required", fieldName)
				case "email":
					msg = fmt.Sprintf("field '%s' must be a valid email address", fieldName)
				case "uuid":
					msg = fmt.Sprintf("field '%s' must be a valid UUID", fieldName)
				default:
					msg = fmt.Sprintf("field '%s' failed validation: %s", fieldName, tag)
				}
				errorMsgs = append(errorMsgs, msg)
			}
			return fmt.Errorf("validation failed: %s", strings.Join(errorMsgs, "; "))
		}
		return fmt.Errorf("validation error: %w", err)
	}

	return nil
}

// BindQuery binds query parameters to struct
func (c *Context) BindQuery(obj interface{}) error {
	return c.Context.ShouldBindQuery(obj)
}

// BindURI binds URI parameters to struct
func (c *Context) BindURI(obj interface{}) error {
	return c.Context.ShouldBindUri(obj)
}

// GetParam returns a URL parameter value
func (c *Context) GetParam(key string) string {
	return c.Context.Param(key)
}

// GetQuery returns a query parameter value
func (c *Context) GetQuery(key string) string {
	return c.Context.Query(key)
}

// GetRawData reads and returns the raw request body
func (c *Context) GetRawData() ([]byte, error) {
	return c.Context.GetRawData()
}

