package worker

import (
	"time"

	"github.com/google/uuid"
	"github.com/vivek6201/biolynq/internal/models"
)

const (
	TaskSendEmail   = "task:send_email"
	TaskRecordEvent = "task:record_event"
)

type SendEmailPayload struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	From    string   `json:"from"`
	Content string   `json:"content"`
}

type RecordEventPayload struct {
	EventType models.EventType `json:"event_type"`
	ProfileID uuid.UUID        `json:"profile_id"`
	LinkID    *uuid.UUID       `json:"link_id,omitempty"`
	IP        string           `json:"ip"`
	UserAgent string           `json:"user_agent"`
	Referrer  string           `json:"referrer"`
	ClickedAt time.Time        `json:"clicked_at"`
}
