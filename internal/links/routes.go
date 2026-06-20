package links

import "github.com/gofiber/fiber/v3"

func RegisterRoute(r fiber.Router, handler ILinkHandler, authMiddleware fiber.Handler) {
	linksRoutes := r.Group("/links", authMiddleware)
	{
		linksRoutes.Get("/", handler.GetAllLinksHandler)
		linksRoutes.Post("/", handler.CreateLinkHandler)
		linksRoutes.Get("/:linkID", handler.GetLinkByIDHandler)
		linksRoutes.Put("/:linkID", handler.UpdateLinkHandler)
		linksRoutes.Delete("/:linkID", handler.DeleteLinkHandler)
	}

	// ShortLinks routes
	shortRoutes := r.Group("/short", authMiddleware)
	{
		shortRoutes.Post("/:linkID", handler.CreateShortLinkHandler)
		shortRoutes.Get("/:linkID", handler.GetShortLinksHandler)
		shortRoutes.Delete("/:linkID/:sid", handler.DeleteShortLinkHandler)
	}
}
