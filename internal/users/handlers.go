package users

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/vivek6201/biolynq/internal/config"
	"github.com/vivek6201/biolynq/internal/utils"
	"gorm.io/gorm"
)

type UserHandler struct {
	service IUserService
	cfg     *config.ConfigVar
}

type IUserHandler interface {
	GetUserHandler(c fiber.Ctx) error
	GetProfileHandler(c fiber.Ctx) error
	UpdateProfileHandler(c fiber.Ctx) error
	GetPublicProfileHandler(c fiber.Ctx) error
}

func NewUserHandler(service IUserService, cfg *config.ConfigVar) IUserHandler {
	return &UserHandler{
		service: service,
		cfg:     cfg,
	}
}

func (h *UserHandler) GetUserHandler(c fiber.Ctx) error {
	userIdStr, ok := c.Locals("userID").(string)
	if !ok || userIdStr == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized access", nil)
	}

	userID, err := uuid.Parse(userIdStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID format in session", err)
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(c, fiber.StatusNotFound, "User not found", nil)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to fetch user", err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "User fetched successfully", user)
}

func (h *UserHandler) GetProfileHandler(c fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized access", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID", nil)
	}

	profile, err := h.service.GetProfile(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(c, fiber.StatusNotFound, "User not found", nil)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to load profile", err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Profile retrieved successfully", profile)
}

func (h *UserHandler) UpdateProfileHandler(c fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized access", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID", nil)
	}

	var profileUpdate UpdateProfileRequest
	if err := c.Bind().Body(&profileUpdate); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if _, err := h.service.UpdateProfile(userID, profileUpdate); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update profile", err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Profile updated successfully", nil)
}

func (h *UserHandler) GetPublicProfileHandler(c fiber.Ctx) error {
	username := c.Params("username")

	profile, err := h.service.GetProfileByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(c, fiber.StatusNotFound, "Profile not found", err)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to fetch profile", err)
	}

	// Capture client metadata and trigger asynchronous profile view event
	ip := getClientIP(c)
	userAgent := c.Get("User-Agent")
	referrer := c.Get("Referer")

	h.service.TrackProfileViewAsync(c.Context(), profile.ID, ip, userAgent, referrer)

	return utils.SendSuccess(c, fiber.StatusOK, "Profile retrieved successfully", profile)
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
