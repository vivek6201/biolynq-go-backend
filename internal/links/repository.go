package links

import (
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/vivek6201/biolynq/internal/models"
	"gorm.io/gorm"
)

type LinkRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

type ILinkRepository interface {
	GetAllLinks(profileID uuid.UUID) ([]LinkResponse, error)
	GetLinkById(id uuid.UUID, profileID uuid.UUID) (*LinkResponse, error)
	CreateLink(link *models.Link) error
	UpdateLink(id uuid.UUID, profileId uuid.UUID, link *UpdateLinkRequest) error
	DeleteLink(id uuid.UUID, profileId uuid.UUID) error
}

func NewLinkRepository(db *gorm.DB, rdb *redis.Client) ILinkRepository {
	return &LinkRepository{
		db:  db,
		rdb: rdb,
	}
}

func (r *LinkRepository) CreateLink(link *models.Link) error {
	return r.db.Create(link).Error
}

func (r *LinkRepository) GetAllLinks(profileId uuid.UUID) ([]LinkResponse, error) {
	var links []LinkResponse

	if err := r.db.Model(&models.Link{}).Where("profile_id = ?", profileId).Order("position ASC").Find(&links).Error; err != nil {
		return nil, err
	}

	return links, nil
}

func (r *LinkRepository) GetLinkById(id uuid.UUID, profileID uuid.UUID) (*LinkResponse, error) {
	var link LinkResponse
	if err := r.db.Model(&models.Link{}).Where("id = ? AND profile_id = ?", id, profileID).First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *LinkRepository) DeleteLink(id uuid.UUID, profileId uuid.UUID) error {
	return r.db.Where("id = ? AND profile_id = ?", id, profileId).Delete(&models.Link{}).Error
}

func (r *LinkRepository) UpdateLink(id uuid.UUID, profileId uuid.UUID, link *UpdateLinkRequest) error {
	return r.db.Model(&models.Link{}).Where("id = ? AND profile_id = ?", id, profileId).Updates(link).Error
}
