// Package models contains the plain data structures shared across the
// application layers.
//
// Models are simple structs tagged for both database (sqlx) and JSON
// serialization. They carry no business logic.
package models

import "time"

// Client represents a client of the service.
type Client struct {
	ID        uint32     `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	Email     string     `db:"email" json:"email"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}
