package users

import "github.com/gofiber/fiber/v3"

func RegisterRoute(r fiber.Router, handler *UserHandler, authMiddleware fiber.Handler) {
	userRoutes := r.Group("/users", authMiddleware)
	{
		userRoutes.Get("/", handler.GetUserHandler)
		userRoutes.Get("/profile", handler.GetProfileHandler)
		userRoutes.Put("/profile", handler.UpdateProfileHandler)
	}

	publicUserRoutes := r.Group("/public")
	{
		publicUserRoutes.Get("/profile/:username", handler.GetPublicProfileHandler)
	}
}
