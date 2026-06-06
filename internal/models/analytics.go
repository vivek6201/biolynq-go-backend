package models

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventTypeProfileView EventType = "profile_view"
	EventTypeLinkClick   EventType = "link_click"
)

type VisitorMetadata struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	IP        string    `json:"ip" gorm:"size:45;not null;uniqueIndex:idx_visitor_client,priority:1"`
	Country   string    `json:"country" gorm:"size:100"`
	City      string    `json:"city" gorm:"size:100"`
	Browser   string    `json:"browser" gorm:"size:100;not null;uniqueIndex:idx_visitor_client,priority:2"`
	OS        string    `json:"os" gorm:"size:100;not null;uniqueIndex:idx_visitor_client,priority:3"`
	Device    string    `json:"device" gorm:"size:100;not null;uniqueIndex:idx_visitor_client,priority:4"`
	UserAgent string    `json:"user_agent" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type AnalyticEvent struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	ProfileID uuid.UUID `json:"profile_id" gorm:"type:uuid;not null;index"`
	Profile   Profile   `json:"-" gorm:"constraint:OnDelete:CASCADE;"`

	EventType EventType `json:"event_type" gorm:"size:50;not null;index"`

	// Nullable foreign key for standard/social links (null for profile_view)
	LinkID *uuid.UUID `json:"link_id" gorm:"type:uuid;index"`
	Link   *Link      `json:"-" gorm:"constraint:OnDelete:CASCADE;"`

	// Non-nullable foreign key for VisitorMetadata
	VisitorMetadataID uuid.UUID       `json:"visitor_metadata_id" gorm:"type:uuid;not null;index"`
	VisitorMetadata   VisitorMetadata `json:"-" gorm:"constraint:OnDelete:CASCADE;"`

	Referrer  string    `json:"referrer" gorm:"type:text"`
	ClickedAt time.Time `json:"clicked_at" gorm:"index"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
