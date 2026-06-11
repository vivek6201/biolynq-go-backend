package server

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/vivek6201/biolynq/internal/config"
	"github.com/vivek6201/biolynq/internal/database"
)

type StructValidator struct {
	validate *validator.Validate
}

func (v *StructValidator) Validate(obj any) error {
	return v.validate.Struct(obj)
}

func StartServer(cfg *config.ConfigVar) {
	db, err := database.ConnectDB(cfg.DB_URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	rdb, err := database.ConnectRedis(cfg.REDIS_URL)

	app := fiber.New(fiber.Config{
		CaseSensitive:  true,
		StrictRouting:  true,
		AppName:        "Biolynq",
		RequestMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"},
		StructValidator: &StructValidator{
			validate: validator.New(),
		},
	})

	app.Use(recover.New())
	app.Use(requestid.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000", "https://app.biolynq.in", "https://biolynq.in"},
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	api := app.Group("/api")
	SetupRoutes(api, db, rdb, cfg)
	if err := app.Listen(":" + cfg.PORT); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
