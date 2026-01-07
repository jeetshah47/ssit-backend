package routes

import (
	subCommands "github.com/equitywala/backend/internal/application/subscription/commands"
	subQueries "github.com/equitywala/backend/internal/application/subscription/queries"
	userCommands "github.com/equitywala/backend/internal/application/user/commands"
	"github.com/equitywala/backend/internal/application/user/queries"
	"github.com/equitywala/backend/internal/config"
	"github.com/equitywala/backend/internal/infrastructure/database/postgres/auth"
	postgresSub "github.com/equitywala/backend/internal/infrastructure/database/postgres/subscription"
	postgresUser "github.com/equitywala/backend/internal/infrastructure/database/postgres/user"
	emailService "github.com/equitywala/backend/internal/infrastructure/email"
	"github.com/equitywala/backend/internal/interfaces/http/api"
	"github.com/equitywala/backend/internal/interfaces/http/handlers"
	"github.com/equitywala/backend/internal/interfaces/http/middleware"
	"github.com/equitywala/backend/internal/shared/jwt"
	"github.com/equitywala/backend/internal/shared/logger"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// Dependencies holds all dependencies needed for routes
type Dependencies struct {
	PostgresDB *gorm.DB
	MongoDB    *mongo.Client
	Config     *config.Config
	Logger     logger.Logger
}

// SetupRoutes configures all application routes
func SetupRoutes(router *gin.Engine, deps *Dependencies) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "equitywala-backend",
		})
	})

	// API v1 routes
	apiGroup := router.Group("/api/v1")
	{
		// Auth routes
		setupAuthRoutes(apiGroup, deps)

		// User routes
		setupUserRoutes(apiGroup, deps)

		// Add more route groups here as modules are implemented
	}
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(apiGroup *gin.RouterGroup, deps *Dependencies) {
	// Initialize repositories
	userRepo := postgresUser.NewRepository(deps.PostgresDB)
	otpRepo := auth.NewOTPRepository(deps.PostgresDB)
	sessionRepo := auth.NewSessionRepository(deps.PostgresDB)
	paymentPlanSelectionRepo := postgresSub.NewRepository(deps.PostgresDB)

	// Initialize command handlers
	createUserCmd := userCommands.NewCreateUserHandler(userRepo)

	// Initialize subscription query handlers
	findPricingPackageQuery := subQueries.NewFindPricingPackageHandler(deps.PostgresDB)

	// Initialize subscription command handlers
	selectPaymentPlanCmd := subCommands.NewSelectPaymentPlanHandler(
		paymentPlanSelectionRepo,
		findPricingPackageQuery,
	)

	// Initialize JWT service
	jwtService := jwt.NewService(
		deps.Config.Auth.GetJWTSecret(),
		deps.Config.Auth.GetJWTExpiry(),
		deps.Config.Auth.GetJWTRefreshExpiry(),
	)

	// Initialize email service
	emailSvc, err := emailService.NewService(deps.Config, deps.Logger)
	if err != nil {
		// Log error but continue - email service is optional for development
		deps.Logger.Warn("Failed to initialize email service", "error", err)
		emailSvc = nil
	}

	// Initialize auth handler
	authHandler := handlers.NewAuthHandler(
		userRepo,
		otpRepo,
		sessionRepo,
		createUserCmd,
		selectPaymentPlanCmd,
		jwtService,
		emailSvc,
	)

	auth := apiGroup.Group("/auth")
	{
		auth.POST("/signup", api.Handle(authHandler.Signup))
		auth.POST("/login", api.Handle(authHandler.Login))
		auth.POST("/verify-otp", api.Handle(authHandler.VerifyOTP))
		auth.POST("/resend-otp", api.Handle(authHandler.ResendOTP))

		// Profile update endpoints (require authentication)
		authProtected := auth.Group("")
		authProtected.Use(middleware.AuthMiddleware(jwtService))
		{
			authProtected.PUT("/profile", api.Handle(authHandler.UpdateProfile))
			authProtected.POST("/verify-pan", api.Handle(authHandler.VerifyPAN))
			authProtected.POST("/select-payment-plan", api.Handle(authHandler.SelectPaymentPlan))
			authProtected.POST("/set-password", api.Handle(authHandler.SetPassword))
		}
	}
}

// setupUserRoutes configures user routes
func setupUserRoutes(apiGroup *gin.RouterGroup, deps *Dependencies) {
	// Initialize repositories
	userRepo := postgresUser.NewRepository(deps.PostgresDB)

	// Initialize query handlers
	getUserQuery := queries.NewGetUserHandler(userRepo)

	// Initialize user handler
	userHandler := handlers.NewUserHandler(getUserQuery)

	// Initialize JWT service for auth middleware
	jwtService := jwt.NewService(
		deps.Config.Auth.GetJWTSecret(),
		deps.Config.Auth.GetJWTExpiry(),
		deps.Config.Auth.GetJWTRefreshExpiry(),
	)

	users := apiGroup.Group("/users")
	users.Use(middleware.AuthMiddleware(jwtService))
	{
		users.GET("/:id", api.Handle(userHandler.GetUser))
	}
}
