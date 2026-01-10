package routes

import (
	"github.com/equitywala/backend/internal/common/jwt"
	"github.com/equitywala/backend/internal/common/logger"
	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/controllers"
	"github.com/equitywala/backend/internal/core/config"
	"github.com/equitywala/backend/internal/core/middlewares"
	emailInfra "github.com/equitywala/backend/internal/infrastructure/email"
	razorpayInfra "github.com/equitywala/backend/internal/infrastructure/razorpay"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/equitywala/backend/internal/services"
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

		// Payment routes
		setupPaymentRoutes(apiGroup, deps)

		// Add more route groups here as modules are implemented
	}
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(apiGroup *gin.RouterGroup, deps *Dependencies) {
	// Initialize repository context
	repoCtx := repositories.NewRepoContext(deps.PostgresDB)

	// Initialize repositories
	userRepo := repositories.NewUserRepo(repoCtx)
	otpRepo := repositories.NewOTPRepo(repoCtx)
	sessionRepo := repositories.NewSessionRepo(repoCtx)
	paymentPlanSelectionRepo := repositories.NewPaymentPlanSelectionRepo(repoCtx)

	// Initialize services
	createUserService := services.NewCreateUserService(userRepo)
	findPricingPackageService := services.NewFindPricingPackageService(deps.PostgresDB)
	selectPaymentPlanService := services.NewSelectPaymentPlanService(
		paymentPlanSelectionRepo,
		findPricingPackageService,
	)

	// Initialize JWT service
	jwtService := jwt.NewService(
		deps.Config.Auth.GetJWTSecret(),
		deps.Config.Auth.GetJWTExpiry(),
		deps.Config.Auth.GetJWTRefreshExpiry(),
	)

	// Initialize email service
	emailSvc, err := emailInfra.NewService(deps.Config, deps.Logger)
	if err != nil {
		// Log error but continue - email service is optional for development
		deps.Logger.Warn("Failed to initialize email service", "error", err)
		emailSvc = nil
	}

	// Initialize auth controller
	authController := controllers.NewAuthController(
		userRepo,
		otpRepo,
		sessionRepo,
		paymentPlanSelectionRepo,
		createUserService,
		selectPaymentPlanService,
		jwtService,
		emailSvc,
	)

	auth := apiGroup.Group("/auth")
	{
		auth.POST("/signup", utils.Handle(authController.Signup))
		auth.POST("/login", utils.Handle(authController.Login))
		auth.POST("/verify-otp", utils.Handle(authController.VerifyOTP))
		auth.POST("/resend-otp", utils.Handle(authController.ResendOTP))
		// Set password can be called without auth (user identified by email, must have verified OTP)
		auth.POST("/set-password", utils.Handle(authController.SetPassword))

		// Profile update endpoints (require authentication)
		authProtected := auth.Group("")
		authProtected.Use(middlewares.AuthMiddleware(jwtService))
		{
			authProtected.PUT("/profile", utils.Handle(authController.UpdateProfile))
			authProtected.POST("/verify-pan", utils.Handle(authController.VerifyPAN))
			authProtected.POST("/select-payment-plan", utils.Handle(authController.SelectPaymentPlan))
		}
	}
}

// setupUserRoutes configures user routes
func setupUserRoutes(apiGroup *gin.RouterGroup, deps *Dependencies) {
	// Initialize repository context
	repoCtx := repositories.NewRepoContext(deps.PostgresDB)

	// Initialize repositories
	userRepo := repositories.NewUserRepo(repoCtx)

	// Initialize services
	getUserService := services.NewGetUserService(userRepo)

	// Initialize user controller
	userController := controllers.NewUserController(getUserService)

	// Initialize JWT service for auth middleware
	jwtService := jwt.NewService(
		deps.Config.Auth.GetJWTSecret(),
		deps.Config.Auth.GetJWTExpiry(),
		deps.Config.Auth.GetJWTRefreshExpiry(),
	)

	users := apiGroup.Group("/users")
	users.Use(middlewares.AuthMiddleware(jwtService))
	{
		users.GET("/:id", utils.Handle(userController.GetUser))
	}
}

// setupPaymentRoutes configures payment routes
func setupPaymentRoutes(apiGroup *gin.RouterGroup, deps *Dependencies) {
	// Initialize repository context
	repoCtx := repositories.NewRepoContext(deps.PostgresDB)

	// Initialize repositories
	paymentRepo := repositories.NewPaymentRepo(repoCtx)
	paymentPlanSelectionRepo := repositories.NewPaymentPlanSelectionRepo(repoCtx)

	// Initialize services
	findPricingPackageService := services.NewFindPricingPackageService(deps.PostgresDB)
	createSubscriptionService := services.NewCreateSubscriptionService(deps.PostgresDB)

	// Initialize Razorpay client
	razorpayClient := razorpayInfra.NewRazorpayClient(deps.Config.Razorpay)

	// Initialize payment service
	paymentService := services.NewPaymentService(
		paymentRepo,
		razorpayClient,
		paymentPlanSelectionRepo,
		findPricingPackageService,
		createSubscriptionService,
		deps.PostgresDB,
	)

	// Initialize payment controller
	paymentController := controllers.NewPaymentController(paymentService)

	// Initialize JWT service for auth middleware
	jwtService := jwt.NewService(
		deps.Config.Auth.GetJWTSecret(),
		deps.Config.Auth.GetJWTExpiry(),
		deps.Config.Auth.GetJWTRefreshExpiry(),
	)

	payments := apiGroup.Group("/payments")
	{
		// Protected routes (require authentication)
		paymentsProtected := payments.Group("")
		paymentsProtected.Use(middlewares.AuthMiddleware(jwtService))
		{
			paymentsProtected.POST("/order", utils.Handle(paymentController.CreateOrder))
			paymentsProtected.POST("/verify", utils.Handle(paymentController.VerifyPayment))
		}

		// Public webhook route (signature verified, no auth required)
		// NOTE: Webhook feature is currently commented out - will be enabled in production
		// payments.POST("/webhook", utils.Handle(paymentController.HandleWebhook))
	}

	// Pricing routes
	setupPricingRoutes(apiGroup, deps)
}

// setupPricingRoutes configures pricing routes
func setupPricingRoutes(apiGroup *gin.RouterGroup, deps *Dependencies) {
	// Initialize services
	listPricingPackagesService := services.NewListPricingPackagesService(deps.PostgresDB)

	// Initialize pricing controller
	pricingController := controllers.NewPricingController(listPricingPackagesService)

	// Public routes (no authentication required)
	pricing := apiGroup.Group("/pricing")
	{
		pricing.GET("/packages", utils.Handle(pricingController.ListPricingPackages))
	}
}
