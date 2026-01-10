package utils

import (
	"net/http"

	"github.com/equitywala/backend/internal/common/errors"
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

	// Check if it's a domain error
	if domainErr, ok := err.(*errors.DomainError); ok {
		statusCode = MapDomainErrorToHTTPStatus(domainErr.Code)
		errorInfo = &ErrorInfo{
			Code:    domainErr.Code,
			Message: domainErr.Message,
		}
	} else {
		// Generic error
		statusCode = http.StatusInternalServerError
		errorInfo = &ErrorInfo{
			Code:    "INTERNAL_ERROR",
			Message: "An internal error occurred",
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
		return err
	}

	// Validate struct tags
	validate := validator.New()
	if err := validate.Struct(obj); err != nil {
		return err
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

