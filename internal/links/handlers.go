package links

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/vivek6201/biolynq/internal/config"
	"github.com/vivek6201/biolynq/internal/utils"
)

type LinkHandler struct {
	service ILinkService
	cfg     *config.ConfigVar
}

type ILinkHandler interface {
	GetAllLinksHandler(ctx fiber.Ctx) error
	CreateLinkHandler(ctx fiber.Ctx) error
	GetLinkByIDHandler(ctx fiber.Ctx) error
	UpdateLinkHandler(ctx fiber.Ctx) error
	DeleteLinkHandler(ctx fiber.Ctx) error

	CreateShortLinkHandler(ctx fiber.Ctx) error
	GetShortLinksHandler(ctx fiber.Ctx) error
	DeleteShortLinkHandler(ctx fiber.Ctx) error
	UpdateShortLinkHandler(ctx fiber.Ctx) error
	CheckSlugHandler(ctx fiber.Ctx) error
}

func NewLinkHandler(service ILinkService, cfg *config.ConfigVar) ILinkHandler {
	return &LinkHandler{
		service: service,
		cfg:     cfg,
	}
}

func (h *LinkHandler) GetAllLinksHandler(ctx fiber.Ctx) error {
	userIDStr, ok := ctx.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendError(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	profileID, err := h.service.GetProfileID(userID)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusNotFound, "Profile not found", err)
	}

	links, err := h.service.GetAllLinks(profileID)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Failed to retrieve links", err)
	}

	baseURL := ctx.BaseURL()
	for i := range links {
		if links[i].ActiveSlug != "" {
			links[i].ShortURL = baseURL + "/s/" + links[i].ActiveSlug
		}
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "Links retrieved successfully", links)
}

func (h *LinkHandler) CreateLinkHandler(ctx fiber.Ctx) error {
	userIDStr, ok := ctx.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendError(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	var req CreateLinkRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid request payload", err)
	}

	profileID, err := h.service.GetProfileID(userID)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusNotFound, "Profile not found", err)
	}

	link, err := h.service.CreateLink(profileID, &req, ctx.BaseURL())
	if err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Failed to create link", err)
	}

	return utils.SendSuccess(ctx, fiber.StatusCreated, "Link created successfully", link)
}

func (h *LinkHandler) GetLinkByIDHandler(ctx fiber.Ctx) error {
	userIDStr, ok := ctx.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendError(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	linkIDStr := ctx.Params("linkID")
	linkID, err := uuid.Parse(linkIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid link ID format", err)
	}

	profileID, err := h.service.GetProfileID(userID)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusNotFound, "Profile not found", err)
	}

	link, err := h.service.GetLinkByID(linkID, profileID)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusNotFound, "Link not found", err)
	}

	if link.ActiveSlug != "" {
		link.ShortURL = ctx.BaseURL() + "/s/" + link.ActiveSlug
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "Link retrieved successfully", link)
}

func (h *LinkHandler) UpdateLinkHandler(ctx fiber.Ctx) error {
	userIDStr, ok := ctx.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendError(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	linkIDStr := ctx.Params("linkID")
	linkID, err := uuid.Parse(linkIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid link ID format", err)
	}

	var req UpdateLinkRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid request payload", err)
	}

	profileID, err := h.service.GetProfileID(userID)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusNotFound, "Profile not found", err)
	}

	err = h.service.UpdateLink(linkID, profileID, &req)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Failed to update link", err)
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "Link updated successfully", nil)
}

func (h *LinkHandler) DeleteLinkHandler(ctx fiber.Ctx) error {
	userIDStr, ok := ctx.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendError(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	linkIDStr := ctx.Params("linkID")
	linkID, err := uuid.Parse(linkIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid link ID format", err)
	}

	profileID, err := h.service.GetProfileID(userID)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusNotFound, "Profile not found", err)
	}

	err = h.service.DeleteLink(linkID, profileID)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Failed to delete link", err)
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "Link deleted successfully", nil)
}

func (h *LinkHandler) CreateShortLinkHandler(ctx fiber.Ctx) error {
	userIDStr, ok := ctx.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendError(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	linkIDStr := ctx.Params("linkID")
	linkID, err := uuid.Parse(linkIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid link ID format", err)
	}

	var req CreateShortLinkRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid request payload", err)
	}

	profileID, err := h.service.GetProfileID(userID)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusNotFound, "Profile not found", err)
	}

	shortLink, err := h.service.CreateShortLink(profileID, linkID, req.Slug, ctx.BaseURL())
	if err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Failed to create short link", err)
	}

	return utils.SendSuccess(ctx, fiber.StatusCreated, "Short link created successfully", shortLink)
}

func (h *LinkHandler) GetShortLinksHandler(ctx fiber.Ctx) error {
	userIDStr, ok := ctx.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendError(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	linkIDStr := ctx.Params("linkID")
	linkID, err := uuid.Parse(linkIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid link ID format", err)
	}

	profileID, err := h.service.GetProfileID(userID)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusNotFound, "Profile not found", err)
	}

	shortLinks, err := h.service.GetShortLinksByLinkID(profileID, linkID, ctx.BaseURL())
	if err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Failed to retrieve short links", err)
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "Short links retrieved successfully", shortLinks)
}

func (h *LinkHandler) DeleteShortLinkHandler(ctx fiber.Ctx) error {
	userIDStr, ok := ctx.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendError(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	linkIDStr := ctx.Params("linkID")
	linkID, err := uuid.Parse(linkIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid link ID format", err)
	}

	slug := ctx.Params("slug")
	if slug == "" {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Slug is required", nil)
	}

	profileID, err := h.service.GetProfileID(userID)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusNotFound, "Profile not found", err)
	}

	err = h.service.DeleteShortLinkBySlug(profileID, linkID, slug)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Failed to delete short link", err)
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "Short link deleted successfully", nil)
}

func (h *LinkHandler) UpdateShortLinkHandler(ctx fiber.Ctx) error {
	userIDStr, ok := ctx.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendError(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	linkIDStr := ctx.Params("linkID")
	linkID, err := uuid.Parse(linkIDStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid link ID format", err)
	}

	slug := ctx.Params("slug")
	if slug == "" {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Slug is required", nil)
	}

	var req UpdateShortLinkRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid request payload", err)
	}

	profileID, err := h.service.GetProfileID(userID)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusNotFound, "Profile not found", err)
	}

	shortLink, err := h.service.UpdateShortLinkBySlug(profileID, linkID, slug, &req, ctx.BaseURL())
	if err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Failed to update short link", err)
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "Short link updated successfully", shortLink)
}

func (h *LinkHandler) CheckSlugHandler(ctx fiber.Ctx) error {
	var req CheckSlugRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Invalid request payload", err)
	}

	available, err := h.service.CheckSlugAvailable(req.Slug)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Failed to check slug availability", err)
	}

	if !available {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Slug is taken", nil)
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "Slug is Available", nil)
}
