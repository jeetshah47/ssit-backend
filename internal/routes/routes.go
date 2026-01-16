package routes

import (
	"github.com/equitywala/backend/internal/common/jwt"
	"github.com/equitywala/backend/internal/common/logger"
	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/controllers"
	"github.com/equitywala/backend/internal/core/config"
	"github.com/equitywala/backend/internal/core/middlewares"
	emailInfra "github.com/equitywala/backend/internal/infrastructure/email"
	paytmInfra "github.com/equitywala/backend/internal/infrastructure/paytm"
	s3Infra "github.com/equitywala/backend/internal/infrastructure/s3"
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

		// Advisory routes
		setupAdvisoryRoutes(apiGroup, deps)

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
		// Log detailed error - email service is required for signup
		deps.Logger.Error("Failed to initialize email service", 
			"error", err,
			"hasAWSAccessKeyID", deps.Config.Email.AWSAccessKeyID != "",
			"hasAWSSecretAccessKey", deps.Config.Email.AWSSecretAccessKey != "",
			"awsRegion", deps.Config.Email.AWSRegion,
			"fromEmail", deps.Config.Email.FromEmail,
			"hasSMTPHost", deps.Config.Email.SMTPHost != "",
		)
		emailSvc = nil
	} else {
		deps.Logger.Info("Email service initialized successfully",
			"awsRegion", deps.Config.Email.AWSRegion,
			"fromEmail", deps.Config.Email.FromEmail,
		)
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
		deps.Logger,
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

	// Initialize Paytm client
	paytmClient := paytmInfra.NewPaytmClient(deps.Config.Paytm)

	// Initialize payment service
	paymentService := services.NewPaymentService(
		paymentRepo,
		paytmClient,
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

// setupAdvisoryRoutes configures advisory routes
func setupAdvisoryRoutes(apiGroup *gin.RouterGroup, deps *Dependencies) {
	// Initialize repository context
	repoCtx := repositories.NewRepoContext(deps.PostgresDB)

	// Initialize repositories
	advisoryTypeRepo := repositories.NewAdvisoryTypeRepo(repoCtx)
	stockBasketRepo := repositories.NewStockBasketRepo(repoCtx)
	ipoAdvisoryRepo := repositories.NewIPOAdvisoryRepo(repoCtx)
	mutualFundBasketRepo := repositories.NewMutualFundBasketRepo(repoCtx)
	sectorSnapshotRepo := repositories.NewSectorSnapshotRepo(repoCtx)

	// Initialize services
	getOverviewService := services.NewGetOverviewService(
		sectorSnapshotRepo,
		stockBasketRepo,
		mutualFundBasketRepo,
	)
	createStockBasketService := services.NewCreateStockBasketService(stockBasketRepo, advisoryTypeRepo)
	getStockBasketService := services.NewGetStockBasketService(stockBasketRepo)
	listStockBasketsService := services.NewListStockBasketsService(stockBasketRepo)
	updateStockBasketService := services.NewUpdateStockBasketService(stockBasketRepo)
	updateStockBasketReportURLService := services.NewUpdateStockBasketReportURLService(stockBasketRepo)

	createIPOAdvisoryService := services.NewCreateIPOAdvisoryService(ipoAdvisoryRepo, advisoryTypeRepo)
	getIPOAdvisoryService := services.NewGetIPOAdvisoryService(ipoAdvisoryRepo)
	listIPOAdvisoriesService := services.NewListIPOAdvisoriesService(ipoAdvisoryRepo)
	updateIPOAdvisoryService := services.NewUpdateIPOAdvisoryService(ipoAdvisoryRepo)
	updateIPOAdvisoryReportURLService := services.NewUpdateIPOAdvisoryReportURLService(ipoAdvisoryRepo)

	createMutualFundBasketService := services.NewCreateMutualFundBasketService(mutualFundBasketRepo, advisoryTypeRepo)
	getMutualFundBasketService := services.NewGetMutualFundBasketService(mutualFundBasketRepo)
	listMutualFundBasketsService := services.NewListMutualFundBasketsService(mutualFundBasketRepo)
	updateMutualFundBasketReportURLService := services.NewUpdateMutualFundBasketReportURLService(mutualFundBasketRepo)

	// Initialize local file storage client for report uploads
	var s3Client services.S3Client
	if deps.Config.S3.BaseDir != "" {
		// Initialize local file storage client
		client, err := s3Infra.NewClient(deps.Config.S3.BaseDir, deps.Config.S3.BaseURL)
		if err != nil {
			deps.Logger.Warn("Failed to initialize file storage client, report uploads will be disabled", "error", err)
			s3Client = nil
		} else {
			s3Client = client
			deps.Logger.Info("File storage client initialized", "baseDir", deps.Config.S3.BaseDir, "baseURL", deps.Config.S3.BaseURL)
		}
	}
	uploadReportService := services.NewUploadAdvisoryReportService(s3Client, deps.Config.S3.BaseDir)
	deleteReportService := services.NewDeleteAdvisoryReportService(s3Client, deps.Config.S3.BaseDir)

	// Initialize controllers
	overviewController := controllers.NewOverviewController(getOverviewService)
	stockBasketController := controllers.NewStockBasketController(
		createStockBasketService,
		getStockBasketService,
		listStockBasketsService,
		updateStockBasketService,
	)
	ipoAdvisoryController := controllers.NewIPOAdvisoryController(
		createIPOAdvisoryService,
		getIPOAdvisoryService,
		listIPOAdvisoriesService,
		updateIPOAdvisoryService,
	)
	mutualFundBasketController := controllers.NewMutualFundBasketController(
		createMutualFundBasketService,
		getMutualFundBasketService,
		listMutualFundBasketsService,
	)
	reportController := controllers.NewAdvisoryReportController(
		uploadReportService,
		deleteReportService,
		updateStockBasketReportURLService,
		updateIPOAdvisoryReportURLService,
		updateMutualFundBasketReportURLService,
	)

	// Initialize JWT service for auth middleware
	jwtService := jwt.NewService(
		deps.Config.Auth.GetJWTSecret(),
		deps.Config.Auth.GetJWTExpiry(),
		deps.Config.Auth.GetJWTRefreshExpiry(),
	)

	// Overview routes (public - requires authentication for user-specific data)
	overview := apiGroup.Group("/overview")
	overview.Use(middlewares.AuthMiddleware(jwtService))
	{
		overview.GET("", utils.Handle(overviewController.GetOverview))
		overview.GET("/sector-snapshots", utils.Handle(overviewController.GetSectorSnapshots))
		overview.GET("/stock-watchlist", utils.Handle(overviewController.GetStockWatchlist))
		overview.GET("/mutual-funds", utils.Handle(overviewController.GetMutualFunds))
	}

	// Advisory routes
	advisory := apiGroup.Group("/advisory")
	{
		// Public read routes (require authentication)
		advisoryProtected := advisory.Group("")
		advisoryProtected.Use(middlewares.AuthMiddleware(jwtService))
		{
			// Stock Baskets
			advisoryProtected.GET("/stock-baskets", utils.Handle(stockBasketController.ListStockBaskets))
			advisoryProtected.GET("/stock-baskets/:id", utils.Handle(stockBasketController.GetStockBasket))

			// IPO Advisories
			advisoryProtected.GET("/ipos", utils.Handle(ipoAdvisoryController.ListIPOAdvisories))
			advisoryProtected.GET("/ipos/:id", utils.Handle(ipoAdvisoryController.GetIPOAdvisory))

			// Mutual Fund Baskets
			advisoryProtected.GET("/mutual-fund-baskets", utils.Handle(mutualFundBasketController.ListMutualFundBaskets))
			advisoryProtected.GET("/mutual-fund-baskets/:id", utils.Handle(mutualFundBasketController.GetMutualFundBasket))
		}

		// Admin/Advisor write routes (require authentication + role check)
		// TODO: Add role-based middleware for Admin/Advisor only
		advisoryAdmin := advisory.Group("")
		advisoryAdmin.Use(middlewares.AuthMiddleware(jwtService))
		{
			// Stock Baskets
			advisoryAdmin.POST("/stock-baskets", utils.Handle(stockBasketController.CreateStockBasket))
			advisoryAdmin.PUT("/stock-baskets/:id", utils.Handle(stockBasketController.UpdateStockBasket))
			advisoryAdmin.POST("/stock-baskets/:id/report", utils.Handle(reportController.UploadStockBasketReport))
			advisoryAdmin.DELETE("/stock-baskets/:id/report", utils.Handle(reportController.DeleteStockBasketReport))

			// IPO Advisories
			advisoryAdmin.POST("/ipos", utils.Handle(ipoAdvisoryController.CreateIPOAdvisory))
			advisoryAdmin.PUT("/ipos/:id", utils.Handle(ipoAdvisoryController.UpdateIPOAdvisory))
			advisoryAdmin.POST("/ipos/:id/report", utils.Handle(reportController.UploadIPOReport))
			advisoryAdmin.DELETE("/ipos/:id/report", utils.Handle(reportController.DeleteIPOReport))

			// Mutual Fund Baskets
			advisoryAdmin.POST("/mutual-fund-baskets", utils.Handle(mutualFundBasketController.CreateMutualFundBasket))
			advisoryAdmin.POST("/mutual-fund-baskets/:id/report", utils.Handle(reportController.UploadMutualFundReport))
			advisoryAdmin.DELETE("/mutual-fund-baskets/:id/report", utils.Handle(reportController.DeleteMutualFundReport))
		}
	}

	// Static file serving for uploaded reports
	if deps.Config.S3.BaseDir != "" {
		apiGroup.Static("/files", deps.Config.S3.BaseDir)
	}
}
