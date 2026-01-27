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
	// Add logger to context middleware
	router.Use(func(c *gin.Context) {
		c.Set("logger", deps.Logger)
		c.Next()
	})

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

	// Initialize repositories for new entities
	etfBasketRepo := repositories.NewETFBasketRepo(repoCtx)
	mfSchemeRepo := repositories.NewMFSchemeRepo(repoCtx)
	nfoRepo := repositories.NewNFORepo(repoCtx)
	webinarRepo := repositories.NewWebinarRepo(repoCtx)
	weeklyMarketMoodRepo := repositories.NewWeeklyMarketMoodRepo(repoCtx)
	weeklyAudioRepo := repositories.NewWeeklyAudioRepo(repoCtx)

	// Initialize services for new entities
	getETFBasketService := services.NewGetETFBasketService(etfBasketRepo)
	listETFBasketsService := services.NewListETFBasketsService(etfBasketRepo)
	getMFSchemeService := services.NewGetMFSchemeService(mfSchemeRepo)
	listMFSchemesService := services.NewListMFSchemesService(mfSchemeRepo)
	getNFOService := services.NewGetNFOService(nfoRepo)
	listNFOsService := services.NewListNFOsService(nfoRepo)
	getWebinarService := services.NewGetWebinarService(webinarRepo)
	listWebinarsService := services.NewListWebinarsService(webinarRepo)
	getWeeklyMarketMoodService := services.NewGetWeeklyMarketMoodService(weeklyMarketMoodRepo)
	getWeeklyAudioService := services.NewGetWeeklyAudioService(weeklyAudioRepo)
	listWeeklyAudioService := services.NewListWeeklyAudioService(weeklyAudioRepo)

	// Initialize controllers
	overviewController := controllers.NewOverviewController(
		getOverviewService,
		sectorSnapshotRepo,
		stockBasketRepo,
		ipoAdvisoryRepo,
		weeklyMarketMoodRepo,
		getWeeklyMarketMoodService,
	)
	stockBasketController := controllers.NewStockBasketController(
		createStockBasketService,
		getStockBasketService,
		listStockBasketsService,
		updateStockBasketService,
		stockBasketRepo,
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
	etfBasketController := controllers.NewETFBasketController(
		getETFBasketService,
		listETFBasketsService,
	)
	mfSchemeController := controllers.NewMFSchemeController(
		getMFSchemeService,
		listMFSchemesService,
	)
	nfoController := controllers.NewNFOController(
		getNFOService,
		listNFOsService,
	)
	webinarController := controllers.NewWebinarController(
		getWebinarService,
		listWebinarsService,
	)
	weeklyAudioController := controllers.NewWeeklyAudioController(
		getWeeklyAudioService,
		listWeeklyAudioService,
	)
	reportController := controllers.NewAdvisoryReportController(
		uploadReportService,
		deleteReportService,
		updateStockBasketReportURLService,
		updateIPOAdvisoryReportURLService,
		updateMutualFundBasketReportURLService,
	)
	adminController := controllers.NewAdminController(
		stockBasketRepo,
		ipoAdvisoryRepo,
		etfBasketRepo,
		mfSchemeRepo,
		mutualFundBasketRepo,
		nfoRepo,
		webinarRepo,
		sectorSnapshotRepo,
		weeklyMarketMoodRepo,
		weeklyAudioRepo,
		advisoryTypeRepo,
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
		overview.GET("/market-pulse", utils.Handle(overviewController.GetMarketPulse))
		overview.GET("/top-call", utils.Handle(overviewController.GetTopCall))
		overview.GET("/weekly-market-mood", utils.Handle(overviewController.GetWeeklyMarketMood))
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
			advisoryProtected.GET("/stock-bullets", utils.Handle(stockBasketController.GetStockBullets))
			advisoryProtected.GET("/stock-recommendations", utils.Handle(stockBasketController.GetStockRecommendations))

			// IPO Advisories
			advisoryProtected.GET("/ipos", utils.Handle(ipoAdvisoryController.ListIPOAdvisories))
			advisoryProtected.GET("/ipos/:id", utils.Handle(ipoAdvisoryController.GetIPOAdvisory))

			// Mutual Fund Baskets
			advisoryProtected.GET("/mutual-fund-baskets", utils.Handle(mutualFundBasketController.ListMutualFundBaskets))
			advisoryProtected.GET("/mutual-fund-baskets/:id", utils.Handle(mutualFundBasketController.GetMutualFundBasket))

			// ETF Baskets
			advisoryProtected.GET("/etfs", utils.Handle(etfBasketController.ListETFBaskets))
			advisoryProtected.GET("/etfs/:id", utils.Handle(etfBasketController.GetETFBasket))

			// MF Schemes
			advisoryProtected.GET("/mf-schemes", utils.Handle(mfSchemeController.ListMFSchemes))
			advisoryProtected.GET("/mf-schemes/:id", utils.Handle(mfSchemeController.GetMFScheme))

			// NFOs
			advisoryProtected.GET("/nfos", utils.Handle(nfoController.ListNFOs))
			advisoryProtected.GET("/nfos/:id", utils.Handle(nfoController.GetNFO))

			// Webinars
			advisoryProtected.GET("/webinars", utils.Handle(webinarController.ListWebinars))
			advisoryProtected.GET("/webinars/:id", utils.Handle(webinarController.GetWebinar))

			// Weekly Audio
			advisoryProtected.GET("/weekly-audio", utils.Handle(weeklyAudioController.GetWeeklyAudio))
			advisoryProtected.GET("/weekly-audio/list", utils.Handle(weeklyAudioController.ListWeeklyAudio))
			advisoryProtected.GET("/weekly-audio/:id", utils.Handle(weeklyAudioController.GetWeeklyAudio))
		}

		// Admin/Advisor write routes (require authentication + role check)
		// TODO: Add role-based middleware for Admin/Advisor only
		advisoryAdmin := advisory.Group("")
		advisoryAdmin.Use(middlewares.AuthMiddleware(jwtService))
		{
			// Stock Baskets
			advisoryAdmin.POST("/stock-baskets", utils.Handle(stockBasketController.CreateStockBasket))
			advisoryAdmin.PUT("/stock-baskets/:id", utils.Handle(stockBasketController.UpdateStockBasket))
			advisoryAdmin.DELETE("/stock-baskets/:id", utils.Handle(stockBasketController.DeleteStockBasket))
			advisoryAdmin.POST("/stock-baskets/:id/report", utils.Handle(reportController.UploadStockBasketReport))
			advisoryAdmin.DELETE("/stock-baskets/:id/report", utils.Handle(reportController.DeleteStockBasketReport))

			// Stock Bullets
			advisoryAdmin.POST("/stock-bullets", utils.Handle(adminController.CreateStockBullet))
			advisoryAdmin.PUT("/stock-bullets/:id", utils.Handle(adminController.UpdateStockBullet))
			advisoryAdmin.DELETE("/stock-bullets/:id", utils.Handle(adminController.DeleteStockBullet))

			// Stock Recommendations
			advisoryAdmin.POST("/stock-recommendations", utils.Handle(adminController.CreateStockRecommendation))
			advisoryAdmin.PUT("/stock-recommendations/:id", utils.Handle(adminController.UpdateStockRecommendation))
			advisoryAdmin.DELETE("/stock-recommendations/:id", utils.Handle(adminController.DeleteStockRecommendation))

			// IPO Advisories
			advisoryAdmin.POST("/ipos", utils.Handle(ipoAdvisoryController.CreateIPOAdvisory))
			advisoryAdmin.PUT("/ipos/:id", utils.Handle(ipoAdvisoryController.UpdateIPOAdvisory))
			advisoryAdmin.DELETE("/ipos/:id", utils.Handle(adminController.DeleteIPO))
			advisoryAdmin.POST("/ipos/:id/report", utils.Handle(reportController.UploadIPOReport))
			advisoryAdmin.DELETE("/ipos/:id/report", utils.Handle(reportController.DeleteIPOReport))

			// ETF Baskets
			advisoryAdmin.POST("/etfs", utils.Handle(adminController.CreateETF))
			advisoryAdmin.PUT("/etfs/:id", utils.Handle(adminController.UpdateETF))
			advisoryAdmin.DELETE("/etfs/:id", utils.Handle(adminController.DeleteETF))

			// MF Schemes
			advisoryAdmin.POST("/mf-schemes", utils.Handle(adminController.CreateMFScheme))
			advisoryAdmin.PUT("/mf-schemes/:id", utils.Handle(adminController.UpdateMFScheme))
			advisoryAdmin.DELETE("/mf-schemes/:id", utils.Handle(adminController.DeleteMFScheme))

			// Mutual Fund Baskets
			advisoryAdmin.POST("/mutual-fund-baskets", utils.Handle(mutualFundBasketController.CreateMutualFundBasket))
			advisoryAdmin.PUT("/mutual-fund-baskets/:id", utils.Handle(adminController.UpdateMFBasket))
			advisoryAdmin.DELETE("/mutual-fund-baskets/:id", utils.Handle(adminController.DeleteMFBasket))
			advisoryAdmin.POST("/mutual-fund-baskets/:id/report", utils.Handle(reportController.UploadMutualFundReport))
			advisoryAdmin.DELETE("/mutual-fund-baskets/:id/report", utils.Handle(reportController.DeleteMutualFundReport))

			// NFOs
			advisoryAdmin.POST("/nfos", utils.Handle(adminController.CreateNFO))
			advisoryAdmin.PUT("/nfos/:id", utils.Handle(adminController.UpdateNFO))
			advisoryAdmin.DELETE("/nfos/:id", utils.Handle(adminController.DeleteNFO))

			// Webinars
			advisoryAdmin.POST("/webinars", utils.Handle(adminController.CreateWebinar))
			advisoryAdmin.PUT("/webinars/:id", utils.Handle(adminController.UpdateWebinar))
			advisoryAdmin.DELETE("/webinars/:id", utils.Handle(adminController.DeleteWebinar))

			// Sectors
			advisoryAdmin.POST("/sector-snapshots", utils.Handle(adminController.CreateSector))
			advisoryAdmin.PUT("/sector-snapshots/:id", utils.Handle(adminController.UpdateSector))
			advisoryAdmin.DELETE("/sector-snapshots/:id", utils.Handle(adminController.DeleteSector))

			// Market Pulse
			advisoryAdmin.POST("/market-pulse", utils.Handle(adminController.CreateMarketPulse))
			advisoryAdmin.PUT("/market-pulse/:id", utils.Handle(adminController.UpdateMarketPulse))
			advisoryAdmin.DELETE("/market-pulse/:id", utils.Handle(adminController.DeleteMarketPulse))

			// Top Call
			advisoryAdmin.PUT("/top-call", utils.Handle(adminController.UpdateTopCall))

			// Weekly Market Mood
			advisoryAdmin.PUT("/weekly-market-mood", utils.Handle(adminController.UpdateWeeklyMarketMood))

			// Weekly Audio
			advisoryAdmin.POST("/weekly-audio", utils.Handle(adminController.UpdateWeeklyAudio))
			advisoryAdmin.PUT("/weekly-audio/:id", utils.Handle(adminController.UpdateWeeklyAudio))
		}
	}

	// Static file serving for uploaded reports
	if deps.Config.S3.BaseDir != "" {
		apiGroup.Static("/files", deps.Config.S3.BaseDir)
	}
}
