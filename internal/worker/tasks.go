package worker

import (
	"time"

	"github.com/google/uuid"
)

const (
	TaskSendEmail   = "task:send_email"
	TaskRecordClick = "task:record_click"
)

type SendEmailPayload struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	From    string   `json:"from"`
	Content string   `json:"content"`
}

type RecordClickPayload struct {
	LinkID    uuid.UUID `json:"link_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Country   string    `json:"country"`
	City      string    `json:"city"`
	Browser   string    `json:"browser"`
	OS        string    `json:"os"`
	Device    string    `json:"device"`
	Referrer  string    `json:"referrer"`
	ClickedAt time.Time `json:"clicked_at"`
}
