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

	// ShortLink repository methods
	CreateShortLink(shortLink *models.ShortLink) error
	GetShortLinksByLinkID(linkID uuid.UUID) ([]models.ShortLink, error)
	DeleteShortLink(shortLinkID uuid.UUID) error
	VerifyLinkOwnership(linkID uuid.UUID, profileID uuid.UUID) (bool, error)
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

	err := r.db.Model(&models.Link{}).
		Select("links.*, COUNT(analytic_events.id) as clicks, active_short_links.slug as active_slug").
		Joins("LEFT JOIN analytic_events ON analytic_events.link_id = links.id AND analytic_events.event_type = ?", models.EventTypeLinkClick).
		Joins("LEFT JOIN short_links as active_short_links ON active_short_links.link_id = links.id AND active_short_links.is_active = ?", true).
		Where("links.profile_id = ?", profileId).
		Group("links.id, active_short_links.slug").
		Order("links.position ASC").
		Scan(&links).Error

	if err != nil {
		return nil, err
	}

	return links, nil
}

func (r *LinkRepository) GetLinkById(id uuid.UUID, profileID uuid.UUID) (*LinkResponse, error) {
	var link LinkResponse
	err := r.db.Model(&models.Link{}).
		Select("links.*, COUNT(analytic_events.id) as clicks, active_short_links.slug as active_slug").
		Joins("LEFT JOIN analytic_events ON analytic_events.link_id = links.id AND analytic_events.event_type = ?", models.EventTypeLinkClick).
		Joins("LEFT JOIN short_links as active_short_links ON active_short_links.link_id = links.id AND active_short_links.is_active = ?", true).
		Where("links.id = ? AND links.profile_id = ?", id, profileID).
		Group("links.id, active_short_links.slug").
		First(&link).Error
	if err != nil {
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

func (r *LinkRepository) CreateShortLink(shortLink *models.ShortLink) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if shortLink.IsActive {
			// Deactivate all other short links for this link
			if err := tx.Model(&models.ShortLink{}).Where("link_id = ?", shortLink.LinkID).Update("is_active", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(shortLink).Error
	})
}

func (r *LinkRepository) GetShortLinksByLinkID(linkID uuid.UUID) ([]models.ShortLink, error) {
	var shortLinks []models.ShortLink
	err := r.db.Where("link_id = ?", linkID).Order("created_at DESC").Find(&shortLinks).Error
	return shortLinks, err
}

func (r *LinkRepository) DeleteShortLink(shortLinkID uuid.UUID) error {
	return r.db.Delete(&models.ShortLink{}, "id = ?", shortLinkID).Error
}

func (r *LinkRepository) VerifyLinkOwnership(linkID uuid.UUID, profileID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Link{}).Where("id = ? AND profile_id = ?", linkID, profileID).Count(&count).Error
	return count > 0, err
}



