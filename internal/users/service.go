package users

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vivek6201/biolynq/internal/cache"
	"github.com/vivek6201/biolynq/internal/models"
	"github.com/vivek6201/biolynq/internal/worker"
)

type UserService struct {
	repo   IUserRepository
	worker worker.TaskDistributor
	caches *Caches // nil-safe cache registry
}

type IUserService interface {
	FindUserByEmail(email string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	CreateUser(user *models.User) error
	CreateTempUser(tempUser *models.TempUser) error
	UpdateTempUser(tempUser *models.TempUser) error
	GetTempUserByID(id uuid.UUID) (*models.TempUser, error)
	GetTempUserByEmail(email string) (*models.TempUser, error)
	CompleteOnboarding(user *models.User, tempUserID uuid.UUID, username string) error
	UpdateProfile(userID uuid.UUID, data UpdateProfileRequest) (*models.Profile, error)
	GetProfile(userID uuid.UUID) (*models.Profile, error)
	GetProfileByUsername(username string) (*models.Profile, error)
	GetSession(sessionID string) (*models.Session, error)
	InvalidateSessionCache(sessionIDs ...string)
	TrackProfileViewAsync(ctx context.Context, profileID uuid.UUID, ip, userAgent, referrer string)
}

// NewUserService constructs a UserService with a pluggable cache registry.
// Pass nil for caches to disable caching entirely (e.g. in the worker).
func NewUserService(
	userRepo IUserRepository,
	worker worker.TaskDistributor,
	caches *Caches,
) IUserService {
	return &UserService{
		repo:   userRepo,
		worker: worker,
		caches: caches,
	}
}

func (s *UserService) FindUserByEmail(email string) (*models.User, error) {
	return s.repo.FindUserByEmail(email)
}

// GetUserByID fetches a user by UUID, using the user cache if available.
func (s *UserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	if s.caches == nil || s.caches.User == nil {
		return s.repo.FindUserById(id)
	}
	key := cache.BuildKey("user", id)
	return s.caches.User.Fetch(context.Background(), key, 1*time.Hour, func() (*models.User, error) {
		return s.repo.FindUserById(id)
	})
}

func (s *UserService) CreateUser(user *models.User) error {
	return s.repo.CreateUser(user)
}

func (s *UserService) CreateTempUser(tempUser *models.TempUser) error {
	return s.repo.CreateTempUser(tempUser)
}

func (s *UserService) UpdateTempUser(tempUser *models.TempUser) error {
	return s.repo.UpdateTempUser(tempUser)
}

func (s *UserService) GetTempUserByID(id uuid.UUID) (*models.TempUser, error) {
	return s.repo.GetTempUserByID(id)
}

func (s *UserService) GetTempUserByEmail(email string) (*models.TempUser, error) {
	return s.repo.GetTempUserByEmail(email)
}

func (s *UserService) CompleteOnboarding(user *models.User, tempUserID uuid.UUID, username string) error {
	return s.repo.CreateUserFromTempUser(user, tempUserID, username)
}

func (s *UserService) UpdateProfile(userID uuid.UUID, data UpdateProfileRequest) (*models.Profile, error) {
	profile, err := s.repo.UpdateProfile(userID, data)
	if err != nil {
		return nil, err
	}

	if s.caches != nil && s.caches.Profile != nil {
		s.caches.Profile.InvalidateAsync(
			cache.BuildKey("profile:user", userID),
			cache.BuildKey("profile:username", profile.Username),
		)
	}

	return profile, nil
}

// GetProfile fetches a profile by user UUID, using the profile cache if available.
func (s *UserService) GetProfile(userID uuid.UUID) (*models.Profile, error) {
	if s.caches == nil || s.caches.Profile == nil {
		return s.repo.GetProfile(userID)
	}
	key := cache.BuildKey("profile:user", userID)
	return s.caches.Profile.Fetch(context.Background(), key, 1*time.Hour, func() (*models.Profile, error) {
		return s.repo.GetProfile(userID)
	})
}

// GetProfileByUsername fetches a profile by username, using the profile cache if available.
func (s *UserService) GetProfileByUsername(username string) (*models.Profile, error) {
	if s.caches == nil || s.caches.Profile == nil {
		return s.repo.GetProfileByUsername(username)
	}
	key := cache.BuildKey("profile:username", username)
	return s.caches.Profile.Fetch(context.Background(), key, 1*time.Hour, func() (*models.Profile, error) {
		return s.repo.GetProfileByUsername(username)
	})
}

// GetSession implements Cache-Aside: check Redis first, then fall back to DB.
func (s *UserService) GetSession(sessionID string) (*models.Session, error) {
	if s.caches == nil || s.caches.Session == nil {
		return s.repo.GetSession(sessionID)
	}
	ctx := context.Background()
	key := cache.BuildKey("session", sessionID)
	return s.caches.Session.Fetch(ctx, key, 24*time.Hour, func() (*models.Session, error) {
		return s.repo.GetSession(sessionID)
	})
}

// InvalidateSessionCache evicts one or more session IDs from the Redis cache.
func (s *UserService) InvalidateSessionCache(sessionIDs ...string) {
	if s.caches == nil || s.caches.Session == nil || len(sessionIDs) == 0 {
		return
	}
	keys := make([]string, len(sessionIDs))
	for i, id := range sessionIDs {
		keys[i] = cache.BuildKey("session", id)
	}
	s.caches.Session.InvalidateAsync(keys...)
}

func (s *UserService) TrackProfileViewAsync(ctx context.Context, profileID uuid.UUID, ip, userAgent, referrer string) {
	payload := &worker.RecordEventPayload{
		EventType: models.EventTypeProfileView,
		ProfileID: profileID,
		IP:        ip,
		UserAgent: userAgent,
		Referrer:  referrer,
		ClickedAt: time.Now(),
	}

	_ = s.worker.DistributeTaskRecordEvent(ctx, payload)
}
