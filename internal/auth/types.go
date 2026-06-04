package auth

import (
	"time"

	"github.com/google/uuid"
)

const (
	EmailOtp = "email_otp"
	Google   = "google"
)

type SendOtpPayload struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyOtpPayload struct {
	Email string `json:"email" validate:"required,email"`
	Otp   string `json:"otp" validate:"required,len=6,numeric"`
}

type CompleteOnBoardingPayload struct {
	TempUserId string `json:"temp_user_id" validate:"required"`
	Username   string `json:"username" validate:"required,min=3,max=30"`
	IPAddress  string `json:"ip_address" validate:"omitempty,ip"`
	UserAgent  string `json:"user_agent" validate:"omitempty"`
}

type CheckUsernamePayload struct {
	Username string `json:"username" validate:"required"`
}

type VerificationResult struct {
	Registered bool
	SessionID  string
	UserID     uuid.UUID
	ExpiresAt  time.Time
	TempUserID uuid.UUID
}

type SessionResult struct {
	SessionID string    `json:"session_id"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}
