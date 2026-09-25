// Package models holds the plain data structures shared across layers.
package models

import "time"

// Client is a signed-up client of the service.
type Client struct {
	ID         uint32     `db:"id" json:"id"`
	Name       string     `db:"name" json:"name"`
	Email      string     `db:"email" json:"email"`
	TelegramID string     `db:"telegram_id" json:"telegram_id"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}
