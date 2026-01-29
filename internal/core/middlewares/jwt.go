package middlewares

import (
	"net/http"
	"strings"

	"github.com/equitywala/backend/internal/common/utils"
	jwtService "github.com/equitywala/backend/internal/common/jwt"
	"github.com/equitywala/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(jwtService *jwtService.Service, roleService *services.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, utils.Response{
				Success: false,
				Error: &utils.ErrorInfo{
					Code:    "UNAUTHORIZED",
					Message: "authorization token required",
				},
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, utils.Response{
				Success: false,
				Error: &utils.ErrorInfo{
					Code:    "INVALID_TOKEN",
					Message: "invalid authorization header format",
				},
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			// Get logger from context if available for logging
			var logger interface{}
			if loggerVal, exists := c.Get("logger"); exists {
				logger = loggerVal
			}

			// Determine specific error type
			var errorCode, errorMessage string
			if err == jwt.ErrTokenExpired {
				errorCode = "TOKEN_EXPIRED"
				errorMessage = "authentication token has expired. Please login again"
			} else if err == jwt.ErrTokenNotValidYet {
				errorCode = "TOKEN_NOT_VALID_YET"
				errorMessage = "authentication token is not yet valid"
			} else if err == jwt.ErrTokenMalformed {
				errorCode = "INVALID_TOKEN"
				errorMessage = "malformed authentication token"
			} else {
				errorCode = "INVALID_TOKEN"
				errorMessage = "invalid authentication token"
			}

			// Log authentication failure
			if logger != nil {
				if l, ok := logger.(interface {
					Warn(msg string, fields ...interface{})
				}); ok {
					l.Warn("Authentication failed",
						"method", c.Request.Method,
						"path", c.Request.URL.Path,
						"error_code", errorCode,
						"error", err.Error(),
						"ip", c.ClientIP(),
					)
				}
			}

			c.JSON(http.StatusUnauthorized, utils.Response{
				Success: false,
				Error: &utils.ErrorInfo{
					Code:    errorCode,
					Message: errorMessage,
				},
			})
			c.Abort()
			return
		}

		// Set user context (will be extracted by utils.NewContext)
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)

		// Fetch user roles from database once; derive primary role from same slice (avoids duplicate DB call)
		userID, err := uuid.Parse(claims.UserID)
		if err == nil && roleService != nil {
			roles, err := roleService.GetUserRoles(c.Request.Context(), userID)
			if err == nil {
				c.Set("roles", roles)
				// Derive primary role from roles (priority: admin > advisor > user) without another DB call
				primaryRole := "user"
				for _, role := range roles {
					if role == "admin" {
						primaryRole = "admin"
						break
					}
					if role == "advisor" {
						primaryRole = "advisor"
					}
				}
				c.Set("role", primaryRole)
			} else {
				c.Set("roles", []string{"user"})
				c.Set("role", "user")
			}
		} else {
			c.Set("roles", []string{"user"})
			c.Set("role", "user")
		}

		c.Next()
	}
}

