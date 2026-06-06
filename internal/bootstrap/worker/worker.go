package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/vivek6201/biolynq/internal/analytics"
	"github.com/vivek6201/biolynq/internal/config"
	"github.com/vivek6201/biolynq/internal/database"
	"github.com/vivek6201/biolynq/internal/users"
	"github.com/vivek6201/biolynq/internal/utils"
	"github.com/vivek6201/biolynq/internal/worker"
)

func StartWorker(cfg *config.ConfigVar) {
	db, err := database.ConnectDB(cfg.DB_URL)
	if err != nil {
		log.Fatalf("Worker: failed to connect to database: %v", err)
	}

	// Dedicated Redis client for the worker (auth repo OTP, session eviction, etc.)
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.REDIS_URL,
	})

	redisOpts := asynq.RedisClientOpt{Addr: cfg.REDIS_URL}
	emailSender := utils.NewEmailSender(cfg.RESEND_KEY)

	// Worker does not serve HTTP so caches are nil — no session/profile
	// lookups happen in the worker hot path.
	geoipService := utils.NewGeoIPService(cfg.GEOIP_DB_PATH)
	userRepo := users.NewUserRepository(db, rdb)
	userService := users.NewUserService(userRepo, nil, nil)
	analyticsRepo := analytics.NewAnalyticsRepository(db)
	analyticsService := analytics.NewAnalyticsService(analyticsRepo, userService, nil, geoipService)

	processor := worker.NewRedisTaskProcessor(redisOpts, emailSender)

	// Register handlers here to prevent circular dependencies between sub-packages
	processor.RegisterHandler(worker.TaskRecordEvent, func(ctx context.Context, task *asynq.Task) error {
		var payload worker.RecordEventPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			log.Printf("Worker: failed to unmarshal payload: %v", err)
			return err
		}

		if err := analyticsService.RecordEvent(
			payload.EventType, payload.ProfileID, payload.LinkID,
			payload.IP, payload.UserAgent, payload.Referrer, payload.ClickedAt,
		); err != nil {
			log.Printf("Worker: failed to record event: %v", err)
			return err
		}

		log.Printf("Worker: recorded event — type=%s profileID=%s", payload.EventType, payload.ProfileID)
		return nil
	})

	log.Println("Starting worker processor...")
	if err := processor.Start(); err != nil {
		log.Fatalf("Worker: processor failed to start: %v", err)
	}
}
