package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/vivek6201/biolynq/internal/users"
	"github.com/vivek6201/biolynq/internal/utils"
	"gorm.io/gorm"
)

// AuthRequired protects routes by validating the user's session
func AuthRequired(userService users.IUserService) fiber.Handler {
	return func(c fiber.Ctx) error {
		var sessionID string

		// 1. Try to retrieve session ID from the session_id Cookie
		sessionID = c.Cookies("session_id")

		// 2. Fall back to Authorization Header (Bearer Token) if Cookie is empty
		if sessionID == "" {
			authHeader := c.Get("Authorization")
			if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
				sessionID = after
			}
		}

		if sessionID == "" {
			return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized: Session token missing", nil)
		}

		// 3. Retrieve the session using the user service
		session, err := userService.GetSession(sessionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized: Session is invalid or expired", nil)
			}
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to authenticate session", err)
		}

		// 4. Inject session details to locals for down-stream handlers
		c.Locals("userID", session.UserID.String())
		c.Locals("sessionID", session.ID)

		return c.Next()
	}
}
