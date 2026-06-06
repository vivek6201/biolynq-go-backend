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
	IsSocial    bool   `json:"is_social" gorm:"default:false;not null"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
