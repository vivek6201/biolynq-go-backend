package links

import "github.com/google/uuid"

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
}

type CreateLinkRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	URL         string `json:"url" validate:"required,url"`
	IconURL     string `json:"icon_url" `
	Position    int    `json:"position" validate:"omitempty,number"`
	IsActive    bool   `json:"is_active" validate:"omitempty,boolean"`
	IsSocial    bool   `json:"is_social" validate:"omitempty,boolean"`
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
