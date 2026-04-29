package routes

import (
	"fmt"
	"net/http"
	"os"
<<<<<<< HEAD
	"stellabill-backend/internal/config"
	"stellabill-backend/internal/cors"
	"stellabill-backend/internal/handlers"
	"stellabill-backend/internal/idempotency"
	"stellabill-backend/internal/middleware"
	"stellabill-backend/internal/repository"
	"stellabill-backend/internal/service"
	"stellabill-backend/internal/startup"
	"stellabill-backend/internal/tracing"

	"stellabill-backend/internal/auth"
	"stellabill-backend/internal/reconciliation"
=======

	"stellarbill-backend/internal/auth"
	"stellarbill-backend/internal/config"
	"stellarbill-backend/internal/handlers"
	"stellarbill-backend/internal/middleware"
	"stellarbill-backend/internal/repository"
	"stellarbill-backend/internal/service"
	"stellarbill-backend/internal/tracing"
	"stellarbill-backend/internal/reconciliation"
	"stellarbill-backend/internal/startup"
>>>>>>> upstream/main

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func Register(r *gin.Engine) {
	cfg, err := config.Load()
	if err != nil {
<<<<<<< HEAD
		log.Printf("Failed to load config: %v", err)
		return
=======
		panic(fmt.Sprintf("failed to load configuration: %v", err))
>>>>>>> upstream/main
	}

	// Initialize tracing
	if cfg.TracingExporter != "none" {
		_, err := tracing.InitTracer(cfg.TracingServiceName)
		if err != nil {
<<<<<<< HEAD
			// Log error but continue
			log.Printf("Failed to initialize tracer: %v", err)
=======
			fmt.Printf("Failed to initialize tracer: %v\n", err)
>>>>>>> upstream/main
		}
	}

	// Global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery())
	r.Use(otelgin.Middleware(cfg.TracingServiceName))
	r.Use(middleware.TraceIDMiddleware())

	// Rate limiting
	rateLimitConfig := middleware.RateLimiterConfig{
		Enabled:        cfg.RateLimitEnabled,
		Mode:           middleware.RateLimitMode(cfg.RateLimitMode),
		RequestsPerSec: int64(cfg.RateLimitRPS),
		BurstSize:      int64(cfg.RateLimitBurst),
		WhitelistPaths: cfg.RateLimitWhitelist,
	}
	r.Use(middleware.RateLimitMiddleware(rateLimitConfig))

	// Request size and Gzip
	r.Use(middleware.RequestSizeLimit(cfg.MaxRequestSize))
	r.Use(middleware.GzipPolicy(middleware.GzipPolicyConfig{
		MaxUncompressedBytes: cfg.MaxGzipUncompressed,
		MaxRatio:             cfg.MaxGzipRatio,
	}))

<<<<<<< HEAD
	subRepo := repository.NewMockSubscriptionRepo(
		&repository.SubscriptionRow{ID: "sub-123", CustomerID: "admin-user", TenantID: "system", Amount: "1999", Currency: "USD"},
		&repository.SubscriptionRow{ID: "sub-456", CustomerID: "merchant-user", TenantID: "merchant-123", Amount: "2999", Currency: "USD"},
		&repository.SubscriptionRow{ID: "test123", CustomerID: "admin-user", TenantID: "system", Amount: "999", Currency: "USD"},
	)
	planRepo := repository.NewMockPlanRepo()
	svc := service.NewSubscriptionService(subRepo, planRepo)

	// Statement service wiring (in-memory mock for test/dev)
=======
	// Dependencies
	subRepo := repository.NewMockSubscriptionRepo()
	_ = repository.NewMockPlanRepo() // Placeholder
>>>>>>> upstream/main
	stmtRepo := repository.NewMockStatementRepo()

	stmtSvc := service.NewStatementService(subRepo, stmtRepo)
	svc := service.NewSubscriptionService(subRepo)
	
	// Create handlers
	h := handlers.NewHandler(nil, nil)
	adminHandler := handlers.NewAdminHandler(cfg.AdminToken)
	
	// Auth configuration
	jwtSecret := cfg.JWTSecret
	// For now, jwksCache is nil as it's not fully wired in config yet
	authMiddleware := middleware.AuthMiddleware(nil, jwtSecret)

	// API Groups
	api := r.Group("/api")
	v1 := api.Group("/v1")
	
	dep := middleware.DeprecationHeaders()

<<<<<<< HEAD
	api.Use(idempotency.Middleware(store))
	{
		// Public health check - no authentication required
		api.GET("/health", dep, handlers.Health)

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			protected.GET("/plans",
				dep,
				auth.RequirePermission(auth.PermReadPlans),
				handlers.ListPlans,
			)

			protected.GET("/subscriptions",
				dep,
				auth.RequirePermission(auth.PermReadSubscriptions),
				handlers.ListSubscriptions,
			)

			protected.GET("/subscriptions/:id",
				dep,
				auth.RequirePermission(auth.PermReadSubscriptions),
				handlers.NewGetSubscriptionHandler(svc),
			)

			protected.GET("/statements/:id", handlers.NewGetStatementHandler(stmtSvc))
			protected.GET("/statements", handlers.NewListStatementsHandler(stmtSvc))
		}

		v1.Use(middleware.AuthMiddleware(jwtSecret))
		{
			v1.GET("/health", handlers.Health)
			v1.GET("/subscriptions", handlers.ListSubscriptions)
			v1.GET("/subscriptions/:id", handlers.NewGetSubscriptionHandler(svc))
			v1.GET("/plans", handlers.ListPlans)
		}

		admin := api.Group("/admin")
		{
			admin.POST("/purge", adminHandler.PurgeCache)
			// Diagnostics endpoint — re-runs startup checks for live triage
			diagHandler := startup.NewDiagnosticsHandler(cfg, nil, nil)
			admin.GET("/diagnostics", auth.RequirePermission(auth.PermManageSubscriptions), diagHandler.Handle)
			// Reconciliation endpoint (admin-only) - accepts backend subscription list
			// Choose adapter implementation via env var CONTRACT_SNAPSHOT_URL. If set, use HTTPAdapter.
			contractURL := os.Getenv("CONTRACT_SNAPSHOT_URL")
			var adapter reconciliation.Adapter
			if contractURL != "" {
				// Optional auth header via CONTRACT_SNAPSHOT_AUTH (e.g. "Bearer <token>")
				authHeader := os.Getenv("CONTRACT_SNAPSHOT_AUTH")
				adapter = reconciliation.NewHTTPAdapter(contractURL, authHeader)
			} else {
				// Default to in-memory adapter (empty) — replace or seed as needed in dev.
				adapter = reconciliation.NewMemoryAdapter()
			}
			// Wire in-memory store for persistence by default; can be swapped for DB-backed store.
			reconStore := reconciliation.NewMemoryStore()
			admin.POST("/reconcile", auth.RequirePermission(auth.PermManageSubscriptions), handlers.NewReconcileHandler(adapter, reconStore))
			// List persisted reports
			admin.GET("/reports", auth.RequirePermission(auth.PermManageSubscriptions), func(c *gin.Context) {
				reports, err := reconStore.ListReports()
				if err != nil {
					c.JSON(500, gin.H{"error": "failed to load reports"})
					return
				}
				c.JSON(200, gin.H{"reports": reports})
			})
		}
=======
	// Public health check
	api.GET("/health", dep, handlers.Health)
	v1.GET("/health", handlers.Health)
	api.GET("/liveness", h.LivenessProbe)
	api.GET("/readiness", h.ReadinessProbe)

	// V1 routes are all protected
	v1.Use(authMiddleware)
	{
		v1.GET("/subscriptions", handlers.ListSubscriptions)
		v1.GET("/subscriptions/:id", handlers.NewGetSubscriptionHandler(svc))
		v1.GET("/plans", handlers.ListPlans)
		v1.GET("/statements/:id", handlers.NewGetStatementHandler(stmtSvc))
		v1.GET("/statements", handlers.NewListStatementsHandler(stmtSvc))
	}

	// Legacy /api routes - also protected
	apiProtected := api.Group("")
	apiProtected.Use(authMiddleware)
	{
		apiProtected.GET("/plans",
			dep,
			auth.RequirePermission(auth.PermReadPlans),
			handlers.ListPlans,
		)

		apiProtected.GET("/subscriptions",
			dep,
			auth.RequirePermission(auth.PermReadSubscriptions),
			handlers.ListSubscriptions,
		)

		apiProtected.GET("/subscriptions/:id",
			dep,
			auth.RequirePermission(auth.PermReadSubscriptions),
			handlers.GetSubscription,
		)

		apiProtected.GET("/statements/:id", handlers.NewGetStatementHandler(stmtSvc))
		apiProtected.GET("/statements", handlers.NewListStatementsHandler(stmtSvc))
	}

	admin := api.Group("/admin")
	admin.Use(authMiddleware)
	{
		admin.POST("/purge", adminHandler.PurgeCache)
		// Diagnostics endpoint — re-runs startup checks for live triage
		diagHandler := startup.NewDiagnosticsHandler(cfg, nil, nil)
		admin.GET("/diagnostics", auth.RequirePermission(auth.PermManageSubscriptions), diagHandler.Handle)
		
		// Reconciliation — scoped by RBAC and tenant
		adapter := reconciliation.NewMemoryAdapter()
		reconStore := reconciliation.NewMemoryStore()
		admin.POST("/reconcile", auth.RequirePermission(auth.PermManageSubscriptions), handlers.NewReconcileHandler(adapter, reconStore))
		admin.GET("/reports", auth.RequirePermission(auth.PermManageSubscriptions), func(c *gin.Context) {
			reports, err := reconStore.ListReports()
			if err != nil {
				c.JSON(500, gin.H{"error": "failed to load reports"})
				return
			}
			c.JSON(200, gin.H{"reports": reports})
		})
>>>>>>> upstream/main
	}
}

