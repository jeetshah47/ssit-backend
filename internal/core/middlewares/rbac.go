package middlewares

import (
	"net/http"

	"github.com/equitywala/backend/internal/common/utils"
	"github.com/gin-gonic/gin"
)

// RoleBasedMiddleware creates a middleware that checks if the user has any of the required roles
func RoleBasedMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get roles from context (set by AuthMiddleware)
		rolesVal, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, utils.Response{
				Success: false,
				Error: &utils.ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "user roles not found in context",
				},
			})
			c.Abort()
			return
		}

		// Type assert roles to []string
		roles, ok := rolesVal.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, utils.Response{
				Success: false,
				Error: &utils.ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "invalid roles format in context",
				},
			})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRequiredRole := false
		for _, userRole := range roles {
			for _, requiredRole := range requiredRoles {
				if userRole == requiredRole {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, utils.Response{
				Success: false,
				Error: &utils.ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "insufficient permissions. required role: " + requiredRoles[0],
				},
			})
			c.Abort()
			return
		}

		// User has required role, continue
		c.Next()
	}
}
