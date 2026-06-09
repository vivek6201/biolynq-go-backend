package auth

import "github.com/gofiber/fiber/v3"

func RegisterRoute(r fiber.Router, handler IAuthHandler, authMiddleware fiber.Handler) {
	authRoutes := r.Group("/auth")
	{
		authRoutes.Post("/otp/send", handler.SendOTPHandler)
		authRoutes.Post("/otp/verify", handler.VerifyOTPHandler)
		authRoutes.Get("/google/login", handler.GoogleLoginHandler)
		authRoutes.Get("/google/callback", handler.GoogleCallbackHandler)
		authRoutes.Post("/register/complete", handler.CompleteRegisterHandler)
		authRoutes.Post("/check-username", handler.CheckUsernameHandler)
		authRoutes.Delete("/logout", authMiddleware, handler.LogoutHandler)
	}
}
