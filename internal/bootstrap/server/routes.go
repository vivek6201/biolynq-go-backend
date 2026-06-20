package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/vivek6201/biolynq/internal/analytics"
	"github.com/vivek6201/biolynq/internal/auth"
	"github.com/vivek6201/biolynq/internal/config"
	"github.com/vivek6201/biolynq/internal/links"
	"github.com/vivek6201/biolynq/internal/middleware"
	"github.com/vivek6201/biolynq/internal/users"
	"github.com/vivek6201/biolynq/internal/utils"
	"github.com/vivek6201/biolynq/internal/worker"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB, rdb *redis.Client, cfg *config.ConfigVar) {
	redisOpts := asynq.RedisClientOpt{Addr: cfg.REDIS_URL}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpts)

	// ── Users ─────────────────────────────────────────────────────────────────
	userRepo := users.NewUserRepository(db, rdb)
	userService := users.NewUserService(userRepo, taskDistributor, users.NewCaches(rdb))
	userHandler := users.NewUserHandler(userService, cfg)

	authMiddleware := middleware.AuthRequired(userService)

	// ── Auth ──────────────────────────────────────────────────────────────────
	authRepo := auth.NewAuthRepository(db, rdb)
	authService := auth.NewAuthService(authRepo, userService, taskDistributor, cfg)
	authHandler := auth.NewAuthHandler(authService, cfg)

	// ── Links ─────────────────────────────────────────────────────────────────
	linksRepo := links.NewLinkRepository(db, rdb)
	linksService := links.NewLinkService(linksRepo, userService, taskDistributor, links.NewCaches(rdb))
	linksHandler := links.NewLinkHandler(linksService, cfg)

	// ── Analytics ─────────────────────────────────────────────────────────────
	geoipService := utils.NewGeoIPService(cfg.GEOIP_DB_PATH)
	analyticsRepo := analytics.NewAnalyticsRepository(db)
	analyticsService := analytics.NewAnalyticsService(analyticsRepo, userService, taskDistributor, geoipService)
	analyticsHandler := analytics.NewAnalyticsHandler(analyticsService, cfg)

	// ── ShortLink Redirect Route at Root ──────────────────────────────────────
	app.Get("/s/:shortId", analyticsHandler.RedirectShortLinkHandler)

	api := app.Group("/api")
	v1 := api.Group("/v1")
	{
		users.RegisterRoute(v1, userHandler, authMiddleware)
		auth.RegisterRoute(v1, authHandler, authMiddleware)
		links.RegisterRoute(v1, linksHandler, authMiddleware)
		analytics.RegisterRoute(v1, analyticsHandler, authMiddleware)
	}
}
