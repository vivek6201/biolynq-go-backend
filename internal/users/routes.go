package users

import "github.com/gofiber/fiber/v3"

func RegisterRoute(r fiber.Router, handler IUserHandler, authMiddleware fiber.Handler) {
	userRoutes := r.Group("/users", authMiddleware)
	{
		userRoutes.Get("/", handler.GetUserHandler)
		userRoutes.Get("/profile", handler.GetProfileHandler)
		userRoutes.Put("/profile", handler.UpdateProfileHandler)
	}

	publicUserRoutes := r.Group("/public")
	{
		publicUserRoutes.Get("/:username", handler.GetPublicProfileHandler)
	}
}
