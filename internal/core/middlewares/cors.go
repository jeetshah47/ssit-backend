package middlewares

import (
	"github.com/equitywala/backend/internal/core/config"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware creates a CORS middleware
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Check if origin is allowed
		allowed := false
		if len(cfg.CORS.AllowedOrigins) == 0 {
			// If no origins specified, allow all (development only)
			allowed = true
		} else {
			for _, allowedOrigin := range cfg.CORS.AllowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

