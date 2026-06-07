package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string    `json:"email" gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"default:null"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type TempUser struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email       string    `json:"email" gorm:"size:255;uniqueIndex;not null"`
	DisplayName string    `json:"display_name" gorm:"size:100"`
	AvatarURL   string    `json:"avatar_url"`
	Provider    string    `json:"provider" gorm:"size:20;default:'email_otp'"`
	IsExpired   bool      `json:"is_expired" gorm:"default:false"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Profile struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	User   User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`

	Username    string `gorm:"size:32;uniqueIndex;not null" json:"username"`
	DisplayName string `gorm:"size:100" json:"display_name"`
	Bio         string `gorm:"type:text" json:"bio"`
	AvatarURL   string `json:"avatar_url"`

	Theme    string `gorm:"size:50;default:'default'" json:"theme"`
	IsPublic bool   `gorm:"default:true" json:"is_public"`

	Links []Link `json:"links,omitempty" gorm:"foreignKey:ProfileID"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Session struct {
	ID string `gorm:"primaryKey;size:255"`

	UserID        uuid.UUID `gorm:"type:uuid;not null;index"`
	User          User      `gorm:"constraint:OnDelete:CASCADE;"`
	Provider      string    `gorm:"size:20;not null"` // "google" or "email_otp"
	IPAddress     string    `gorm:"size:45"`          // Supports IPv4 and IPv6
	UserAgent     string    `gorm:"type:text"`
	ExpiresAt     time.Time `gorm:"not null;index"`
	LastRotatedAt time.Time `gorm:"not null"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
}
