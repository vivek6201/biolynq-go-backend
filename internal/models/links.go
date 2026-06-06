package models

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	ProfileID uuid.UUID `json:"profile_id" gorm:"type:uuid;not null"`
	Profile   Profile   `json:"-" gorm:"constraint:OnDelete:CASCADE;"`

	Title       string `json:"title" gorm:"size:255;not null"`
	Description string `json:"description" gorm:"type:text"`
	URL         string `json:"url" gorm:"size:255;not null"`
	IconURL     string `json:"icon_url" gorm:"type:text;not null"`
	Position    int    `json:"position" gorm:"default:0;not null"`
	IsActive    bool   `json:"is_active" gorm:"default:true;not null"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type LinkStats struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	LinkID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	Link   Link      `gorm:"constraint:OnDelete:CASCADE;"`

	TotalClicks int64 `gorm:"default:0"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type LinkClickEvent struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	LinkID uuid.UUID `json:"link_id" gorm:"type:uuid;not null;index"`
	Link   Link      `json:"-" gorm:"constraint:OnDelete:CASCADE;"`

	Country string `json:"country" gorm:"size:100"`
	City    string `json:"city" gorm:"size:100"`

	Browser string `json:"browser" gorm:"size:100"`
	OS      string `json:"os" gorm:"size:100"`
	Device  string `json:"device" gorm:"size:100"`

	Referrer  string    `json:"referrer" gorm:"type:text"`
	ClickedAt time.Time `json:"clicked_at" gorm:"index"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type SocialLink struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	ProfileID uuid.UUID `json:"profile_id" gorm:"type:uuid;not null;index"`
	Profile   Profile   `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	Platform  string    `json:"platform" gorm:"size:50"`
	URL       string    `json:"url" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
