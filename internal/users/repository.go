package users

import (
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/vivek6201/biolynq/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

type IUserRepository interface {
	FindUserByEmail(email string) (*models.User, error)
	FindUserById(id uuid.UUID) (*models.User, error)
	CreateUser(user *models.User) error
	CreateTempUser(tempUser *models.TempUser) error
	UpdateTempUser(tempUser *models.TempUser) error
	GetTempUserByID(id uuid.UUID) (*models.TempUser, error)
	GetTempUserByEmail(email string) (*models.TempUser, error)
	CreateUserFromTempUser(user *models.User, tempUserID uuid.UUID, username string) error
	UpdateProfile(userID uuid.UUID, data UpdateProfileRequest) (*models.Profile, error)
	GetProfile(userID uuid.UUID) (*models.Profile, error)
	GetProfileByID(id uuid.UUID) (*models.Profile, error)
	GetProfileByUsername(username string) (*models.Profile, error)
	GetSession(sessionID string) (*models.Session, error)
}

func NewUserRepository(db *gorm.DB, rdb *redis.Client) IUserRepository {
	return &UserRepository{
		db:  db,
		rdb: rdb,
	}
}

func (r *UserRepository) FindUserByEmail(email string) (*models.User, error) {
	var user *models.User

	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) FindUserById(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) CreateTempUser(tempUser *models.TempUser) error {
	return r.db.Create(tempUser).Error
}

func (r *UserRepository) UpdateTempUser(tempUser *models.TempUser) error {
	return r.db.Save(tempUser).Error
}

func (r *UserRepository) GetTempUserByID(id uuid.UUID) (*models.TempUser, error) {
	var tempUser models.TempUser
	if err := r.db.Where("id = ? AND is_expired = ?", id, false).First(&tempUser).Error; err != nil {
		return nil, err
	}
	return &tempUser, nil
}

func (r *UserRepository) GetTempUserByEmail(email string) (*models.TempUser, error) {
	var tempUser models.TempUser
	if err := r.db.Where("email = ?", email).First(&tempUser).Error; err != nil {
		return nil, err
	}
	return &tempUser, nil
}

func (r *UserRepository) CreateUserFromTempUser(user *models.User, tempUserID uuid.UUID, username string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		var tempUser models.TempUser
		if err := tx.First(&tempUser, "id = ?", tempUserID).Error; err != nil {
			return err
		}

		profile := models.Profile{
			UserID:      user.ID,
			Username:    username,
			DisplayName: tempUser.DisplayName,
			AvatarURL:   tempUser.AvatarURL,
			Theme:       "default",
			IsPublic:    true,
		}

		if err := tx.Create(&profile).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.TempUser{}).Where("id = ?", tempUserID).Update("is_expired", true).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *UserRepository) UpdateProfile(userID uuid.UUID, data UpdateProfileRequest) (*models.Profile, error) {
	var profile models.Profile
	if err := r.db.Model(&profile).Where("user_id = ?", userID).Updates(data).Error; err != nil {
		return nil, err
	}

	if err := r.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *UserRepository) GetProfile(userID uuid.UUID) (*models.Profile, error) {
	var profile models.Profile
	if err := r.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *UserRepository) GetProfileByID(id uuid.UUID) (*models.Profile, error) {
	var profile models.Profile
	if err := r.db.Where("id = ?", id).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *UserRepository) GetProfileByUsername(username string) (*models.Profile, error) {
	var profile models.Profile
	err := r.db.Preload("Links", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_active = ?", true).Order("position ASC")
	}).Where("username = ?", username).First(&profile).Error

	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *UserRepository) GetSession(sessionID string) (*models.Session, error) {
	var session models.Session
	if err := r.db.Where("id = ? AND expires_at > ?", sessionID, time.Now()).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}
