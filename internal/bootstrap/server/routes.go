package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/vivek6201/biolynq/internal/analytics"
	"github.com/vivek6201/biolynq/internal/auth"
	"github.com/vivek6201/biolynq/internal/config"
	"github.com/vivek6201/biolynq/internal/links"
	"github.com/vivek6201/biolynq/internal/users"
	"github.com/vivek6201/biolynq/internal/worker"
	"gorm.io/gorm"
)

func SetupRoutes(r fiber.Router, db *gorm.DB, rdb *redis.Client, cfg *config.ConfigVar) {
	redisOpts := asynq.RedisClientOpt{
		Addr: cfg.REDIS_URL,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpts)

	userRepo := users.NewUserRepository(db, rdb)
	userService := users.NewUserService(userRepo)
	userHandler := users.NewUserHandler(userService, cfg)

	authRepo := auth.NewAuthRepository(db, rdb)
	authService := auth.NewAuthService(authRepo, userService, taskDistributor, cfg)
	authHandler := auth.NewAuthHandler(authService, cfg)

	v1 := r.Group("v1")
	{
		users.RegisterRoute(v1, userHandler)
		auth.RegisterRoute(v1, authHandler)
		links.RegisterRoute(v1)
		analytics.RegisterRoute(v1)
	}
}
