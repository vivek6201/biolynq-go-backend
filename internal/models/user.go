package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string    `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string    `gorm:"default:null"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type TempUser struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email       string    `gorm:"size:255;uniqueIndex;not null"`
	DisplayName string    `gorm:"size:100"`
	AvatarURL   string
	Provider    string `gorm:"size:20;default:'email_otp'"` // "google" or "email_otp"
	IsExpired   bool   `gorm:"default:false"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Profile struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	User   User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`

	Username    string `gorm:"size:32;uniqueIndex;not null" json:"username"`
	DisplayName string `gorm:"size:100" json:"display_name"`
	Bio         string `gorm:"type:text" json:"bio"`
	AvatarURL   string `json:"avatar_url"`

	Theme    string `gorm:"size:50;default:'default'" json:"theme"`
	IsPublic bool   `gorm:"default:true" json:"is_public"`

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
