// Package models contains the plain data structures shared across layers.
//
// Models carry no business logic and use both db and json struct tags;
// pointer fields mark nullable columns.
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
