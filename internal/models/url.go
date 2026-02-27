package models

import (
	"time"
)

type URL struct {
	ID          uint
	OriginalURL string
	ShortCode   string
	CreateBy    string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	IsActive    bool
}

type ShortCode string

type ClickEvent struct {
	ShortCode ShortCode
	ClickedAt time.Time
	IPAddress string
	Referrer  string
	UserAgent string
}
