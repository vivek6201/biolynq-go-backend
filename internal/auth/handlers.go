package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/vivek6201/biolynq/internal/config"
	"github.com/vivek6201/biolynq/internal/utils"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	service IAuthService
	cfg     *config.ConfigVar
}

type IAuthHandler interface {
	SendOTPHandler(c fiber.Ctx) error
	VerifyOTPHandler(c fiber.Ctx) error
	GoogleLoginHandler(c fiber.Ctx) error
	GoogleCallbackHandler(c fiber.Ctx) error
	CompleteRegisterHandler(c fiber.Ctx) error
	LogoutHandler(c fiber.Ctx) error
	CheckUsernameHandler(c fiber.Ctx) error
}

func NewAuthHandler(service IAuthService, cfg *config.ConfigVar) IAuthHandler {
	return &AuthHandler{
		service: service,
		cfg:     cfg,
	}
}

func (h *AuthHandler) SendOTPHandler(c fiber.Ctx) error {
	var req SendOtpPayload
	if err := c.Bind().JSON(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request payload", err)
	}

	err := h.service.SendOtp(req.Email)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Failed to send OTP", err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "OTP sent successfully", nil)
}

// Verify OTP and generate session token

func (h *AuthHandler) VerifyOTPHandler(c fiber.Ctx) error {
	var req VerifyOtpPayload
	if err := c.Bind().JSON(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request payload", err)
	}

	ip := c.IP()
	ua := c.Get("User-Agent")

	result, err := h.service.VerifyOtp(req.Email, req.Otp, ip, ua)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Verification failed", err)
	}

	if !result.Registered {
		return utils.SendSuccess(c, fiber.StatusOK, "Email verified, onboarding pending", fiber.Map{
			"registered":   false,
			"temp_user_id": result.TempUserID,
		})
	}

	// Set HTTP-only Cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    result.SessionID,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false, // Set to true if deploying over HTTPS
		Expires:  result.ExpiresAt,
	})

	return utils.SendSuccess(c, fiber.StatusOK, "Login successful", fiber.Map{
		"registered": true,
		"session_id": result.SessionID,
		"expires_at": result.ExpiresAt,
	})
}

// Google Login Handler

func (h *AuthHandler) GoogleLoginHandler(c fiber.Ctx) error {
	// Generate random OAuth state to prevent CSRF
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	state := hex.EncodeToString(stateBytes)

	// Save state in secure short-lived cookie (5 mins)
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(5 * time.Minute),
	})

	redirectURL := h.service.GetGoogleAuthURL(state)
	return c.Redirect().To(redirectURL)
}

// Google Callback Handler

func (h *AuthHandler) GoogleCallbackHandler(c fiber.Ctx) error {
	state := c.Query("state")
	cookieState := c.Cookies("oauth_state")

	// Validate CSRF state
	if state == "" || state != cookieState {
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid OAuth state configuration", nil)
	}

	// Delete state cookie
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(-1 * time.Hour),
	})

	code := c.Query("code")
	if code == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Google authorization code missing", nil)
	}

	ip := c.IP()
	ua := c.Get("User-Agent")

	result, err := h.service.HandleGoogleCallback(code, ip, ua)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Google authentication failed", err)
	}

	if !result.Registered {
		// New user signup: Redirect to select username on frontend
		redirectURL := h.cfg.FRONTEND_URL + "/get-started?temp_user_id=" + result.TempUserID.String() + "&registered=false"
		return c.Redirect().To(redirectURL)
	}

	// Existing user login: Set Session ID Cookie and redirect to dashboard
	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    result.SessionID,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		Expires:  result.ExpiresAt,
	})

	redirectURL := h.cfg.FRONTEND_URL + "/get-started?session_id=" + result.SessionID + "&registered=true"
	return c.Redirect().To(redirectURL)

}

// Complete Register Handler

func (h *AuthHandler) CompleteRegisterHandler(c fiber.Ctx) error {
	var req CompleteOnBoardingPayload
	if err := c.Bind().JSON(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request payload", err)
	}

	tempUserID, err := uuid.Parse(req.TempUserId)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid temp_user_id format", err)
	}

	payload := CompleteOnBoardingPayload{
		TempUserId: tempUserID.String(),
		Username:   req.Username,
		IPAddress:  c.IP(),
		UserAgent:  c.Get("User-Agent"),
	}

	result, err := h.service.CompleteOnBoarding(payload)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Registration failed", err)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    result.SessionID,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		Expires:  result.ExpiresAt,
	})

	return utils.SendSuccess(c, fiber.StatusOK, "Registration onboarding completed", fiber.Map{
		"session_id": result.SessionID,
		"expires_at": result.ExpiresAt,
	})
}

// Logout Handler

func (h *AuthHandler) LogoutHandler(c fiber.Ctx) error {
	sessionID, ok := c.Locals("sessionID").(string)
	if !ok || sessionID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "No active session found", nil)
	}

	err := h.service.RevokeSession(sessionID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to revoke session", err)
	}

	// Clear Cookie explicitly with same properties used to set it
	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(-24 * time.Hour),
	})

	return utils.SendSuccess(c, fiber.StatusOK, "Logged out successfully", nil)
}

func (h *AuthHandler) CheckUsernameHandler(c fiber.Ctx) error {
	var body CheckUsernamePayload

	if err := c.Bind().JSON(&body); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request payload", err)
	}

	result, err := h.service.CheckUsername(body.Username)
	if result {
		return utils.SendError(c, fiber.StatusBadRequest, "Username is taken", nil)
	}

	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Failed to check username", err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Username is Available", nil)
}
