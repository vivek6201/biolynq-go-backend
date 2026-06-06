package links

import (
	"github.com/google/uuid"
	"github.com/vivek6201/biolynq/internal/models"
	"github.com/vivek6201/biolynq/internal/users"
	"github.com/vivek6201/biolynq/internal/worker"
)

type LinkService struct {
	repo        ILinkRepository
	userService users.IUserService
	worker      worker.TaskDistributor
}

type ILinkService interface {
	GetProfileID(userID uuid.UUID) (uuid.UUID, error)
	GetAllLinks(profileID uuid.UUID) ([]LinkResponse, error)
	GetLinkByID(id uuid.UUID, profileID uuid.UUID) (*LinkResponse, error)
	CreateLink(profileID uuid.UUID, req *CreateLinkRequest) (*LinkResponse, error)
	UpdateLink(id uuid.UUID, profileID uuid.UUID, req *UpdateLinkRequest) error
	DeleteLink(id uuid.UUID, profileID uuid.UUID) error
}

func NewLinkService(repo ILinkRepository, userService users.IUserService, worker worker.TaskDistributor) ILinkService {
	return &LinkService{
		repo:        repo,
		userService: userService,
		worker:      worker,
	}
}

func (s *LinkService) GetProfileID(userID uuid.UUID) (uuid.UUID, error) {
	profile, err := s.userService.GetProfile(userID)
	if err != nil {
		return uuid.Nil, err
	}
	return profile.ID, nil
}

func (s *LinkService) GetAllLinks(profileID uuid.UUID) ([]LinkResponse, error) {
	return s.repo.GetAllLinks(profileID)
}

func (s *LinkService) GetLinkByID(id uuid.UUID, profileID uuid.UUID) (*LinkResponse, error) {
	return s.repo.GetLinkById(id, profileID)
}

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

	return &LinkResponse{
		ID:          link.ID,
		Title:       link.Title,
		Description: link.Description,
		URL:         link.URL,
		IconURL:     link.IconURL,
		Position:    link.Position,
		IsActive:    link.IsActive,
		IsSocial:    link.IsSocial,
	}, nil
}

func (s *LinkService) UpdateLink(id uuid.UUID, profileID uuid.UUID, req *UpdateLinkRequest) error {
	return s.repo.UpdateLink(id, profileID, req)
}

func (s *LinkService) DeleteLink(id uuid.UUID, profileID uuid.UUID) error {
	return s.repo.DeleteLink(id, profileID)
}
