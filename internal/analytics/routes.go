package analytics

import "github.com/gofiber/fiber/v3"

func RegisterRoute(r fiber.Router, handler IAnalyticsHandler, authMiddleware fiber.Handler) {
	// Public redirection endpoint
	publicRoutes := r.Group("/public")
	{
		publicRoutes.Get("/links/:linkID/redirect", handler.RedirectLinkHandler)
	}

	// Authenticated analytics dashboard endpoints
	analyticsRoutes := r.Group("/analytics", authMiddleware)
	{
		analyticsRoutes.Get("/overview", handler.GetOverviewHandler)
		analyticsRoutes.Get("/links", handler.GetLinksStatsHandler)
		analyticsRoutes.Get("/breakdown", handler.GetDemographicsHandler)
	}
}
