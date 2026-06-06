package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vivek6201/biolynq/internal/config"
	"github.com/vivek6201/biolynq/internal/models"
	"github.com/vivek6201/biolynq/internal/templates"
	"github.com/vivek6201/biolynq/internal/users"
	"github.com/vivek6201/biolynq/internal/worker"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type AuthService struct {
	repo            IAuthRepository
	userService     users.IUserService
	taskDistributor worker.TaskDistributor
	oauthConfig     *oauth2.Config
}

// define structure of service
type IAuthService interface {
	SendOtp(email string) error
	VerifyOtp(email string, otp string, ip string, ua string) (*VerificationResult, error)
	GetGoogleAuthURL(state string) string
	HandleGoogleCallback(code string, ip string, ua string) (*VerificationResult, error)
	CompleteOnBoarding(payload CompleteOnBoardingPayload) (*SessionResult, error)
	CheckUsername(username string) (bool, error)
	RevokeSession(sessionID string) error
}

func NewAuthService(repo IAuthRepository, userService users.IUserService, distributor worker.TaskDistributor, cfg *config.ConfigVar) IAuthService {
	return &AuthService{
		repo:            repo,
		userService:     userService,
		taskDistributor: distributor,
		oauthConfig: &oauth2.Config{
			ClientID:     cfg.GOOGLE_CLIENT_ID,
			ClientSecret: cfg.GOOGLE_CLIENT_SECRET,
			RedirectURL:  cfg.GOOGLE_REDIRECT_URL,
			Endpoint:     google.Endpoint,
			Scopes: []string{
				"openid",
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
		},
	}
}

func (s *AuthService) SendOtp(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))

	otp, err := generateOTPCode()
	if err != nil {
		return fmt.Errorf("failed to generate otp: %w", err)
	}

	err = s.repo.StoreOTP(email, otp, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to store otp: %w", err)
	}

	htmlContent, err := templates.RenderTemplate("otp.html", map[string]string{
		"OTP": otp,
	})
	if err != nil {
		return fmt.Errorf("failed to render otp template: %w", err)
	}

	payload := &worker.SendEmailPayload{
		To:      []string{email},
		From:    "onboarding@biolynq.in",
		Subject: "Verify Your Email Address",
		Content: htmlContent,
	}

	err = s.taskDistributor.DistributeTaskSendEmail(context.TODO(), payload)
	if err != nil {
		return fmt.Errorf("failed to distribute email task: %w", err)
	}
	return nil
}

func (s *AuthService) VerifyOtp(email string, otp string, ip string, ua string) (*VerificationResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	otp = strings.TrimSpace(otp)

	ok, err := s.repo.VerifyOTP(email, otp)
	if err != nil {
		return nil, fmt.Errorf("failed to verify otp: %w", err)
	}

	if !ok {
		return nil, fmt.Errorf("invalid otp: %s", otp)
	}

	return s.handleUserVerification(email, EmailOtp, ip, ua)
}

func (s *AuthService) GetGoogleAuthURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *AuthService) HandleGoogleCallback(code string, ip string, ua string) (*VerificationResult, error) {
	// Exchange code for token
	token, err := s.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// fetch user details from google userinfo
	profile, err := fetchGoogleUserProfile(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch google userinfo: %w", err)
	}

	email := strings.TrimSpace(strings.ToLower(profile.Email))

	return s.handleUserVerification(email, Google, ip, ua)
}

// handleUserVerification handles the registration/login flow common to OTP and OAuth
func (s *AuthService) handleUserVerification(email string, provider string, ip string, ua string) (*VerificationResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	// Check if user already exists
	user, err := s.userService.FindUserByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// User doesn't exist in DB -> Create/Get Temp User
			var tempUser *models.TempUser
			t, dbErr := s.userService.GetTempUserByEmail(email)
			if dbErr != nil {
				if errors.Is(dbErr, gorm.ErrRecordNotFound) {
					// Create temp user
					newTempUser := &models.TempUser{
						ID:        uuid.New(),
						Email:     email,
						Provider:  provider,
						IsExpired: false,
					}
					if err := s.userService.CreateTempUser(newTempUser); err != nil {
						return nil, err
					}
					tempUser = newTempUser
				} else {
					return nil, dbErr
				}
			} else {
				t.IsExpired = false
				t.Provider = provider
				if err := s.userService.UpdateTempUser(t); err != nil {
					return nil, err
				}
				tempUser = t
			}

			return &VerificationResult{
				Registered: false,
				TempUserID: tempUser.ID,
			}, nil
		}
		return nil, err
	}

	// User exists -> Create Session
	sessionId, err := generateRandomSessionToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	session := &models.Session{
		ID:            sessionId,
		UserID:        user.ID,
		Provider:      provider,
		IPAddress:     ip,
		UserAgent:     ua,
		ExpiresAt:     expiresAt,
		LastRotatedAt: time.Now(),
	}

	evictedIDs, err := s.repo.CreateSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	if len(evictedIDs) > 0 {
		s.userService.InvalidateSessionCache(evictedIDs...)
	}

	return &VerificationResult{
		Registered: true,
		SessionID:  sessionId,
		UserID:     user.ID,
		ExpiresAt:  expiresAt,
	}, nil
}

func (s *AuthService) CompleteOnBoarding(payload CompleteOnBoardingPayload) (*SessionResult, error) {
	username := strings.TrimSpace(strings.ToLower(payload.Username))

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

	if !usernameRegex.MatchString(username) {
		return nil, fmt.Errorf("invalid username format")
	}

	tempUserID, err := uuid.Parse(payload.TempUserId)
	if err != nil {
		return nil, fmt.Errorf("invalid temp user ID format: %w", err)
	}

	tempUser, err := s.userService.GetTempUserByID(tempUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("onboarding session not found or has expired")
		}
		return nil, err
	}

	// Verify username uniqueness
	_, err = s.userService.GetProfileByUsername(username)
	if err == nil {
		return nil, errors.New("username is already taken")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	user := &models.User{
		ID:    uuid.New(),
		Email: tempUser.Email,
	}

	if err := s.userService.CompleteOnboarding(user, tempUserID, username); err != nil {
		return nil, fmt.Errorf("failed to complete onboarding: %w", err)
	}

	// Create session
	sessionId, err := generateRandomSessionToken()
	if err != nil {
		return nil, err
	}

	var expiresAt time.Time
	if tempUser.Provider == "google" {
		expiresAt = time.Now().Add(30 * 24 * time.Hour) // 30 days
	} else {
		expiresAt = time.Now().Add(24 * time.Hour) // 1 day
	}

	session := &models.Session{
		ID:            sessionId,
		UserID:        user.ID,
		Provider:      tempUser.Provider,
		IPAddress:     payload.IPAddress,
		UserAgent:     payload.UserAgent,
		ExpiresAt:     expiresAt,
		LastRotatedAt: time.Now(),
	}

	evictedIDs, err := s.repo.CreateSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	if len(evictedIDs) > 0 {
		s.userService.InvalidateSessionCache(evictedIDs...)
	}

	return &SessionResult{
		SessionID: sessionId,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *AuthService) RevokeSession(sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}
	if err := s.repo.DeleteSession(sessionID); err != nil {
		return err
	}
	// Evict from cache at the service layer — repository only handles DB
	s.userService.InvalidateSessionCache(sessionID)
	return nil
}

func (s *AuthService) CheckUsername(username string) (bool, error) {
	username = strings.TrimSpace(username)

	// Validate username format (regex: alphanumeric + underscores, 3-30 chars)
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)
	if !usernameRegex.MatchString(username) {
		return false, errors.New("username must be 3-30 characters long and contain only letters, numbers, and underscores")
	}

	// Verify username uniqueness
	profile, err := s.userService.GetProfileByUsername(username)
	if profile != nil {
		return true, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	return false, err
}
