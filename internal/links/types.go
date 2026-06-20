package links

import (
	"time"

	"github.com/google/uuid"
)

type LinkResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	IconURL     string    `json:"icon_url"`
	Position    int       `json:"position"`
	IsActive    bool      `json:"is_active"`
	IsSocial    bool      `json:"is_social"`
	Clicks      int64     `json:"clicks"`
	ShortURL    string    `json:"short_url,omitempty"`
	ActiveSlug  string    `json:"active_slug,omitempty"`
}

type CreateLinkRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	URL         string `json:"url" validate:"required,url"`
	IconURL     string `json:"icon_url" `
	Position    int    `json:"position" validate:"omitempty,number"`
	IsActive    bool   `json:"is_active" validate:"omitempty,boolean"`
	IsSocial    bool   `json:"is_social" validate:"omitempty,boolean"`
	Shorten     *bool  `json:"shorten" validate:"omitempty,boolean"`
	ShortAlias  string `json:"short_alias" validate:"omitempty,min=3,max=50"`
}

type UpdateLinkRequest struct {
	Title       *string `json:"title" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
	URL         *string `json:"url" validate:"omitempty,url"`
	IconURL     *string `json:"icon_url" validate:"omitempty"`
	Position    *int    `json:"position" validate:"omitempty,number"`
	IsActive    *bool   `json:"is_active" validate:"omitempty,boolean"`
	IsSocial    *bool   `json:"is_social" validate:"omitempty,boolean"`
}

type ShortLinkResponse struct {
	ID        uuid.UUID `json:"id"`
	LinkID    uuid.UUID `json:"link_id"`
	Slug      string    `json:"slug"`
	ShortURL  string    `json:"short_url"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateShortLinkRequest struct {
	Slug string `json:"slug" validate:"omitempty,min=3,max=50"`
}

