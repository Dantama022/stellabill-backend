package routes

import (
	"log"
	"os"
	"stellarbill-backend/internal/config"
	"stellarbill-backend/internal/cors"
	"stellarbill-backend/internal/handlers"
	"stellarbill-backend/internal/idempotency"
	"stellarbill-backend/internal/middleware"
	"stellarbill-backend/internal/repository"
	"stellarbill-backend/internal/service"

	"stellarbill-backend/internal/auth"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	corsProfile := cors.ProfileForEnv(cfg.Env, cfg.AllowedOrigins)

	// Apply rate limiting middleware
	rateLimitConfig := middleware.RateLimiterConfig{
		Enabled:        cfg.RateLimitEnabled,
		Mode:           middleware.RateLimitMode(cfg.RateLimitMode),
		RequestsPerSec: int64(cfg.RateLimitRPS),
		BurstSize:      int64(cfg.RateLimitBurst),
		WhitelistPaths: cfg.RateLimitWhitelist,
	}
	r.Use(middleware.RateLimitMiddleware(rateLimitConfig))

	r.Use(cors.Middleware(corsProfile))

	store := idempotency.NewStore(idempotency.DefaultTTL)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret"
	}

	subRepo := repository.NewMockSubscriptionRepo()
	planRepo := repository.NewMockPlanRepo()
	svc := service.NewSubscriptionService(subRepo, planRepo)
	// wire planRepo into handlers for list/detail endpoints and optional caching
	handlers.SetPlanRepository(planRepo)

	// Define the API version/group
	api := r.Group("/api")
	v1 := api.Group("/v1")

	dep := middleware.DeprecationHeaders()

	api.Use(idempotency.Middleware(store))
	v1.Use(middleware.AuthMiddleware(jwtSecret))
	{
		// Public health check - no authentication required
		api.GET("/health", dep, handlers.Health)
		v1.GET("/health", handlers.Health)

		// Public read (user + admin)
		api.GET("/plans",
			dep,
			auth.RequirePermission(auth.PermReadPlans),
			handlers.ListPlans,
		)

		api.GET("/subscriptions",
			dep,
			auth.RequirePermission(auth.PermReadSubscriptions),
			handlers.ListSubscriptions,
		)

		api.GET("/subscriptions/:id",
			dep,
			auth.RequirePermission(auth.PermReadSubscriptions),
			handlers.GetSubscription,
		)

		// Example future admin-only endpoints:
		// api.POST("/plans", auth.RequirePermission(auth.PermManagePlans), ...)
		api.GET("/subscriptions", dep, handlers.ListSubscriptions)
		v1.GET("/subscriptions", handlers.ListSubscriptions)
		api.GET("/subscriptions/:id", dep, middleware.AuthMiddleware(jwtSecret), handlers.NewGetSubscriptionHandler(svc))
		v1.GET("/subscriptions/:id", middleware.AuthMiddleware(jwtSecret), handlers.NewGetSubscriptionHandler(svc))
		api.GET("/plans", dep, handlers.ListPlans)
		v1.GET("/plans", handlers.ListPlans)

		api.GET("/statements/:id", dep, middleware.AuthMiddleware(jwtSecret), handlers.NewGetStatementHandler(stmtSvc))
		v1.GET("/statements/:id", middleware.AuthMiddleware(jwtSecret), handlers.NewGetStatementHandler(stmtSvc))
		api.GET("/statements", dep, middleware.AuthMiddleware(jwtSecret), handlers.NewListStatementsHandler(stmtSvc))
		v1.GET("/statements", middleware.AuthMiddleware(jwtSecret), handlers.NewListStatementsHandler(stmtSvc))

		admin := api.Group("/admin")
		admin.POST("/purge", dep, adminHandler.PurgeCache)

		adminV1 := v1.Group("/admin")
		adminV1.POST("/purge", adminHandler.PurgeCache)
	}
}
