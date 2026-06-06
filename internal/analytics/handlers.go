package analytics

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/vivek6201/biolynq/internal/config"
	"github.com/vivek6201/biolynq/internal/utils"
	"gorm.io/gorm"
)

type AnalyticsHandler struct {
	service IAnalyticsService
	cfg     *config.ConfigVar
}

type IAnalyticsHandler interface {
	GetOverviewHandler(c fiber.Ctx) error
	GetLinksStatsHandler(c fiber.Ctx) error
	GetDemographicsHandler(c fiber.Ctx) error
	RedirectLinkHandler(c fiber.Ctx) error
}

func NewAnalyticsHandler(service IAnalyticsService, cfg *config.ConfigVar) IAnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
		cfg:     cfg,
	}
}

func (h *AnalyticsHandler) GetOverviewHandler(c fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	overview, err := h.service.GetOverview(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(c, fiber.StatusNotFound, "Profile not found", err)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to load overview analytics", err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Overview analytics retrieved successfully", overview)
}

func (h *AnalyticsHandler) GetLinksStatsHandler(c fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	stats, err := h.service.GetLinkStats(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(c, fiber.StatusNotFound, "Profile not found", err)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to load link statistics", err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Link statistics retrieved successfully", stats)
}

func (h *AnalyticsHandler) GetDemographicsHandler(c fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	demographics, err := h.service.GetDemographics(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(c, fiber.StatusNotFound, "Profile not found", err)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to load demographics analytics", err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Demographics analytics retrieved successfully", demographics)
}

func (h *AnalyticsHandler) RedirectLinkHandler(c fiber.Ctx) error {
	linkIDStr := c.Params("linkID")
	linkID, err := uuid.Parse(linkIDStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid link ID format", err)
	}

	link, err := h.service.GetLinkByID(linkID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(c, fiber.StatusNotFound, "Link not found", err)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to resolve redirect link", err)
	}

	// Capture client metadata
	ip := getClientIP(c)
	userAgent := c.Get("User-Agent")
	referrer := c.Get("Referer")

	// Trigger asynchronous click tracking task via service
	h.service.TrackClickAsync(c.Context(), link, ip, userAgent, referrer)

	// Perform 302 Found redirect
	return c.Redirect().To(link.URL)
}

func getClientIP(c fiber.Ctx) string {
	forwarded := c.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	realIP := c.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	return c.IP()
}
