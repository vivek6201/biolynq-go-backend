package links

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vivek6201/biolynq/internal/cache"
	"github.com/vivek6201/biolynq/internal/models"
	"github.com/vivek6201/biolynq/internal/users"
	"github.com/vivek6201/biolynq/internal/utils"
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
	CreateLink(profileID uuid.UUID, req *CreateLinkRequest, baseURL string) (*LinkResponse, error)
	UpdateLink(id uuid.UUID, profileID uuid.UUID, req *UpdateLinkRequest) error
	DeleteLink(id uuid.UUID, profileID uuid.UUID) error

	// ShortLink endpoints
	CreateShortLink(profileID uuid.UUID, linkID uuid.UUID, slug string, baseURL string) (*ShortLinkResponse, error)
	GetShortLinksByLinkID(profileID uuid.UUID, linkID uuid.UUID, baseURL string) ([]ShortLinkResponse, error)
	DeleteShortLink(profileID uuid.UUID, linkID uuid.UUID, shortLinkID uuid.UUID) error
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
func (s *LinkService) CreateLink(profileID uuid.UUID, req *CreateLinkRequest, baseURL string) (*LinkResponse, error) {
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

	var activeSlug string
	if req.Shorten != nil && *req.Shorten {
		shortLinkResponse, err := s.CreateShortLink(profileID, link.ID, req.ShortAlias, baseURL)
		if err == nil {
			activeSlug = shortLinkResponse.Slug
		}
	}

	if s.caches != nil && s.caches.Links != nil {
		s.caches.Links.InvalidateAsync(cache.BuildKey("links:profile", profileID))
	}

	s.userService.InvalidateProfileCacheByProfileID(profileID)

	shortURL := ""
	if activeSlug != "" {
		shortURL = baseURL + "/s/" + activeSlug
	}

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
		ShortURL:    shortURL,
		ActiveSlug:  activeSlug,
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

func (s *LinkService) CreateShortLink(profileID uuid.UUID, linkID uuid.UUID, slug string, baseURL string) (*ShortLinkResponse, error) {
	owned, err := s.repo.VerifyLinkOwnership(linkID, profileID)
	if err != nil {
		return nil, err
	}
	if !owned {
		return nil, errors.New("unauthorized: link does not belong to profile")
	}

	if slug == "" {
		generated, err := utils.GenerateRandomString(8)
		if err != nil {
			return nil, err
		}
		slug = generated
	}

	shortLink := &models.ShortLink{
		ID:       uuid.New(),
		LinkID:   linkID,
		Slug:     slug,
		IsActive: true,
	}

	if err := s.repo.CreateShortLink(shortLink); err != nil {
		return nil, err
	}

	if s.caches != nil && s.caches.Links != nil {
		s.caches.Links.InvalidateAsync(cache.BuildKey("links:profile", profileID))
	}
	s.userService.InvalidateProfileCacheByProfileID(profileID)

	return &ShortLinkResponse{
		ID:        shortLink.ID,
		LinkID:    shortLink.LinkID,
		Slug:      shortLink.Slug,
		ShortURL:  baseURL + "/s/" + shortLink.Slug,
		IsActive:  shortLink.IsActive,
		CreatedAt: shortLink.CreatedAt,
	}, nil
}

func (s *LinkService) GetShortLinksByLinkID(profileID uuid.UUID, linkID uuid.UUID, baseURL string) ([]ShortLinkResponse, error) {
	owned, err := s.repo.VerifyLinkOwnership(linkID, profileID)
	if err != nil {
		return nil, err
	}
	if !owned {
		return nil, errors.New("unauthorized: link does not belong to profile")
	}

	dbShortLinks, err := s.repo.GetShortLinksByLinkID(linkID)
	if err != nil {
		return nil, err
	}

	responses := make([]ShortLinkResponse, len(dbShortLinks))
	for i, sl := range dbShortLinks {
		responses[i] = ShortLinkResponse{
			ID:        sl.ID,
			LinkID:    sl.LinkID,
			Slug:      sl.Slug,
			ShortURL:  baseURL + "/s/" + sl.Slug,
			IsActive:  sl.IsActive,
			CreatedAt: sl.CreatedAt,
		}
	}
	return responses, nil
}

func (s *LinkService) DeleteShortLink(profileID uuid.UUID, linkID uuid.UUID, shortLinkID uuid.UUID) error {
	owned, err := s.repo.VerifyLinkOwnership(linkID, profileID)
	if err != nil {
		return err
	}
	if !owned {
		return errors.New("unauthorized: link does not belong to profile")
	}

	if err := s.repo.DeleteShortLink(shortLinkID); err != nil {
		return err
	}

	if s.caches != nil && s.caches.Links != nil {
		s.caches.Links.InvalidateAsync(cache.BuildKey("links:profile", profileID))
	}
	s.userService.InvalidateProfileCacheByProfileID(profileID)

	return nil
}

