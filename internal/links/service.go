package links

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vivek6201/biolynq/internal/cache"
	"github.com/vivek6201/biolynq/internal/models"
	"github.com/vivek6201/biolynq/internal/users"
	"github.com/vivek6201/biolynq/internal/worker"
)

type LinkService struct {
	repo        ILinkRepository
	userService users.IUserService
	worker      worker.TaskDistributor
	caches      *Caches // nil-safe cache registry
}

type ILinkService interface {
	GetProfileID(userID uuid.UUID) (uuid.UUID, error)
	GetAllLinks(profileID uuid.UUID) ([]LinkResponse, error)
	GetLinkByID(id uuid.UUID, profileID uuid.UUID) (*LinkResponse, error)
	CreateLink(profileID uuid.UUID, req *CreateLinkRequest) (*LinkResponse, error)
	UpdateLink(id uuid.UUID, profileID uuid.UUID, req *UpdateLinkRequest) error
	DeleteLink(id uuid.UUID, profileID uuid.UUID) error
}

// NewLinkService constructs a LinkService with a pluggable cache registry.
// Pass nil for caches to disable caching entirely.
func NewLinkService(
	repo ILinkRepository,
	userService users.IUserService,
	worker worker.TaskDistributor,
	caches *Caches,
) ILinkService {
	return &LinkService{
		repo:        repo,
		userService: userService,
		worker:      worker,
		caches:      caches,
	}
}

func (s *LinkService) GetProfileID(userID uuid.UUID) (uuid.UUID, error) {
	profile, err := s.userService.GetProfile(userID)
	if err != nil {
		return uuid.Nil, err
	}
	return profile.ID, nil
}

// GetAllLinks fetches all links for a profile, using the list cache if available.
func (s *LinkService) GetAllLinks(profileID uuid.UUID) ([]LinkResponse, error) {
	if s.caches == nil || s.caches.Links == nil {
		return s.repo.GetAllLinks(profileID)
	}
	key := cache.BuildKey("links:profile", profileID)
	result, err := s.caches.Links.Fetch(context.Background(), key, 5*time.Minute, func() (*[]LinkResponse, error) {
		links, err := s.repo.GetAllLinks(profileID)
		if err != nil {
			return nil, err
		}
		return &links, nil
	})
	if err != nil {
		return nil, err
	}
	return *result, nil
}

// GetLinkByID fetches a single link, using the single-link cache if available.
func (s *LinkService) GetLinkByID(id uuid.UUID, profileID uuid.UUID) (*LinkResponse, error) {
	if s.caches == nil || s.caches.Link == nil {
		return s.repo.GetLinkById(id, profileID)
	}
	key := cache.BuildKey("link", id)
	return s.caches.Link.Fetch(context.Background(), key, 5*time.Minute, func() (*LinkResponse, error) {
		return s.repo.GetLinkById(id, profileID)
	})
}

// CreateLink creates a link and invalidates the list cache asynchronously.
func (s *LinkService) CreateLink(profileID uuid.UUID, req *CreateLinkRequest) (*LinkResponse, error) {
	link := &models.Link{
		ID:          uuid.New(),
		ProfileID:   profileID,
		Title:       req.Title,
		Description: req.Description,
		URL:         req.URL,
		IconURL:     req.IconURL,
		Position:    req.Position,
		IsActive:    req.IsActive,
		IsSocial:    req.IsSocial,
	}

	if err := s.repo.CreateLink(link); err != nil {
		return nil, err
	}

	if s.caches != nil && s.caches.Links != nil {
		s.caches.Links.InvalidateAsync(cache.BuildKey("links:profile", profileID))
	}

	s.userService.InvalidateProfileCacheByProfileID(profileID)

	return &LinkResponse{
		ID:          link.ID,
		Title:       link.Title,
		Description: link.Description,
		URL:         link.URL,
		IconURL:     link.IconURL,
		Position:    link.Position,
		IsActive:    link.IsActive,
		IsSocial:    link.IsSocial,
		Clicks:      0,
	}, nil
}

// UpdateLink updates a link in the DB and invalidates both caches asynchronously.
func (s *LinkService) UpdateLink(id uuid.UUID, profileID uuid.UUID, req *UpdateLinkRequest) error {
	if err := s.repo.UpdateLink(id, profileID, req); err != nil {
		return err
	}

	if s.caches != nil {
		if s.caches.Link != nil {
			s.caches.Link.InvalidateAsync(cache.BuildKey("link", id))
		}
		if s.caches.Links != nil {
			s.caches.Links.InvalidateAsync(cache.BuildKey("links:profile", profileID))
		}
	}

	s.userService.InvalidateProfileCacheByProfileID(profileID)
	return nil
}

// DeleteLink removes a link from DB and evicts both caches asynchronously.
func (s *LinkService) DeleteLink(id uuid.UUID, profileID uuid.UUID) error {
	if err := s.repo.DeleteLink(id, profileID); err != nil {
		return err
	}

	if s.caches != nil {
		if s.caches.Link != nil {
			s.caches.Link.InvalidateAsync(cache.BuildKey("link", id))
		}
		if s.caches.Links != nil {
			s.caches.Links.InvalidateAsync(cache.BuildKey("links:profile", profileID))
		}
	}

	s.userService.InvalidateProfileCacheByProfileID(profileID)
	return nil
}
