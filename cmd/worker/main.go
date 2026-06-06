package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/vivek6201/biolynq/internal/analytics"
	"github.com/vivek6201/biolynq/internal/config"
	"github.com/vivek6201/biolynq/internal/database"
	"github.com/vivek6201/biolynq/internal/users"
	"github.com/vivek6201/biolynq/internal/utils"
	"github.com/vivek6201/biolynq/internal/worker"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file:", err)
	}

	cfg := config.LoadConfig()

	db, err := database.ConnectDB(cfg.DB_URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	redisOpts := asynq.RedisClientOpt{
		Addr: cfg.REDIS_URL,
	}

	emailSender := utils.NewEmailSender(cfg.RESEND_KEY)

	// Initialize dependencies for analytics service
	geoipService := utils.NewGeoIPService(cfg.GEOIP_DB_PATH)
	userRepo := users.NewUserRepository(db, nil)
	userService := users.NewUserService(userRepo, nil)
	analyticsRepo := analytics.NewAnalyticsRepository(db)
	analyticsService := analytics.NewAnalyticsService(analyticsRepo, userService, nil, geoipService)

	processor := worker.NewRedisTaskProcessor(redisOpts, emailSender)

	// Register TaskRecordEvent handler dynamically to prevent circular dependencies in sub-packages
	processor.RegisterHandler(worker.TaskRecordEvent, func(ctx context.Context, task *asynq.Task) error {
		var payload worker.RecordEventPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			log.Printf("Worker Error: failed to unmarshal payload: %v", err)
			return err
		}

		err := analyticsService.RecordEvent(payload.EventType, payload.ProfileID, payload.LinkID, payload.IP, payload.UserAgent, payload.Referrer, payload.ClickedAt)
		if err != nil {
			log.Printf("Worker Error: failed to record event in database: %v", err)
			return err
		}

		log.Printf("Worker: Successfully recorded event to DB: Type=%s, ProfileID=%s", payload.EventType, payload.ProfileID)
		return nil
	})

	log.Println("Starting worker processor...")
	if err := processor.Start(); err != nil {
		log.Fatalf("Failed to start worker processor: %v", err)
	}
}
