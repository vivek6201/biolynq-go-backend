package main

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/vivek6201/biolynq/internal/config"
	"github.com/vivek6201/biolynq/internal/database"
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

	processor := worker.NewRedisTaskProcessor(redisOpts, emailSender, db)

	log.Println("Starting worker processor...")
	if err := processor.Start(); err != nil {
		log.Fatalf("Failed to start worker processor: %v", err)
	}
}
