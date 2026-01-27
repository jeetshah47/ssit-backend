package middlewares

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/equitywala/backend/internal/common/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiterConfig defines rate limiting configuration
type RateLimiterConfig struct {
	// Requests per second
	RPS float64
	// Burst size (allows short bursts above RPS)
	Burst int
	// Cleanup interval for old entries
	CleanupInterval time.Duration
	// Entry expiration time
	Expiration time.Duration
}

// Default rate limit configurations
var (
	// Strict rate limit for OTP operations (5 requests per minute, burst of 2)
	StrictRateLimit = RateLimiterConfig{
		RPS:             5.0 / 60.0, // 5 per minute = 0.083 per second
		Burst:           2,
		CleanupInterval: 5 * time.Minute,
		Expiration:      15 * time.Minute,
	}

	// Moderate rate limit for login (10 requests per minute, burst of 3)
	ModerateRateLimit = RateLimiterConfig{
		RPS:             10.0 / 60.0, // 10 per minute = 0.167 per second
		Burst:           3,
		CleanupInterval: 5 * time.Minute,
		Expiration:      15 * time.Minute,
	}

	// Standard rate limit for general auth endpoints (20 requests per minute, burst of 5)
	StandardRateLimit = RateLimiterConfig{
		RPS:             20.0 / 60.0, // 20 per minute = 0.333 per second
		Burst:           5,
		CleanupInterval: 5 * time.Minute,
		Expiration:      15 * time.Minute,
	}
)

// ipLimiter holds rate limiter for a specific IP
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter manages rate limiting per IP address
type RateLimiter struct {
	limiters map[string]*ipLimiter
	mu       sync.RWMutex
	config   RateLimiterConfig
	stop     chan struct{}
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*ipLimiter),
		config:   config,
		stop:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// getLimiter returns the rate limiter for the given IP, creating one if needed
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = &ipLimiter{
			limiter:  rate.NewLimiter(rate.Limit(rl.config.RPS), rl.config.Burst),
			lastSeen: time.Now(),
		}
		rl.limiters[ip] = limiter
	} else {
		limiter.lastSeen = time.Now()
	}

	return limiter.limiter
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, limiter := range rl.limiters {
				if now.Sub(limiter.lastSeen) > rl.config.Expiration {
					delete(rl.limiters, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.stop:
			return
		}
	}
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stop)
}

// getClientIP extracts the client IP address from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs separated by commas
		// The first IP is the original client IP
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			// Trim whitespace from the first IP
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}
	// Check X-Real-IP header
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	// Fallback to RemoteAddr
	return c.ClientIP()
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)
		lim := limiter.getLimiter(ip)

		if !lim.Allow() {
			c.JSON(http.StatusTooManyRequests, utils.Response{
				Success: false,
				Error: &utils.ErrorInfo{
					Code:    "RATE_LIMIT_EXCEEDED",
					Message: "Too many requests. Please try again later.",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CreateRateLimitMiddleware creates a rate limiting middleware with the given configuration
func CreateRateLimitMiddleware(config RateLimiterConfig) gin.HandlerFunc {
	limiter := NewRateLimiter(config)
	return RateLimitMiddleware(limiter)
}
