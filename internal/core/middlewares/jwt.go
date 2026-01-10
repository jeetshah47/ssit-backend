package middlewares

import (
	"net/http"
	"strings"

	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/common/jwt"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(jwtService *jwt.Service) gin.HandlerFunc {
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
			c.JSON(http.StatusUnauthorized, utils.Response{
				Success: false,
				Error: &utils.ErrorInfo{
					Code:    "INVALID_TOKEN",
					Message: "invalid or expired token",
				},
			})
			c.Abort()
			return
		}

		// Set user context (will be extracted by utils.NewContext)
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		// Role can be extracted from user if needed, for now set default
		c.Set("role", "user")

		c.Next()
	}
}

