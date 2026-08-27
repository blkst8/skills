// Package models defines plain data structs with both db and json tags.
// Models carry no business logic; pointer types mark nullable columns.
package models

import "time"

type Invoice struct {
	ID         uint32     `db:"id" json:"id"`
	CustomerID uint32     `db:"customer_id" json:"customer_id"`
	Number     string     `db:"number" json:"number"`
	Amount     float64    `db:"amount" json:"amount"`
	Currency   string     `db:"currency" json:"currency"`
	Status     string     `db:"status" json:"status"`
	DueDate    time.Time  `db:"due_date" json:"due_date"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}
