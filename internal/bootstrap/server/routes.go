package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/vivek6201/biolynq/internal/analytics"
	"github.com/vivek6201/biolynq/internal/auth"
	"github.com/vivek6201/biolynq/internal/config"
	"github.com/vivek6201/biolynq/internal/links"
	"github.com/vivek6201/biolynq/internal/users"
	"gorm.io/gorm"
)

func SetupRoutes(r fiber.Router, db *gorm.DB, cfg *config.ConfigVar) {
	userRepo := users.NewUserRepository(db)
	userService := users.NewUserService(userRepo)
	userHandler := users.NewUserHandler(userService, cfg)

	v1 := r.Group("v1")
	{
		users.RegisterRoute(v1, userHandler)
		auth.RegisterRoute(v1)
		links.RegisterRoute(v1)
		analytics.RegisterRoute(v1)
	}
}
