package users

import (
	"github.com/google/uuid"
	"github.com/vivek6201/biolynq/internal/models"
)

type UserService struct {
	repo *UserRepository
}

func NewUserService(userRepo *UserRepository) *UserService {
	return &UserService{
		repo: userRepo,
	}
}

func (s *UserService) FindUserByEmail(email string) (*models.User, error) {
	return s.repo.FindUserByEmail(email)
}

func (s *UserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.repo.FindUserById(id)
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
	return s.repo.UpdateProfile(userID, data)
}

func (s *UserService) GetProfile(userID uuid.UUID) (*models.Profile, error) {
	return s.repo.GetProfile(userID)
}

func (s *UserService) GetProfileByUsername(username string) (*models.Profile, error) {
	return s.repo.GetProfileByUsername(username)
}

func (s *UserService) GetSession(sessionID string) (*models.Session, error) {
	return s.repo.GetSession(sessionID)
}
