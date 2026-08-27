// Package models contains the plain data structures shared across layers.
//
// Models carry no business logic; they only declare db and json tags.
package models

import "time"

// InvoiceStatus is the lifecycle state of an invoice.
type InvoiceStatus string

const (
	InvoiceStatusPending InvoiceStatus = "pending"
	InvoiceStatusPaid    InvoiceStatus = "paid"
	InvoiceStatusOverdue InvoiceStatus = "overdue"
)

// Invoice is a single customer invoice.
type Invoice struct {
	ID          int64         `db:"id" json:"id"`
	Number      string        `db:"number" json:"number"`
	CustomerID  int64         `db:"customer_id" json:"customer_id"`
	AmountCents int64         `db:"amount_cents" json:"amount_cents"`
	Currency    string        `db:"currency" json:"currency"`
	Status      InvoiceStatus `db:"status" json:"status"`
	DueAt       time.Time     `db:"due_at" json:"due_at"`
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt   *time.Time    `db:"updated_at" json:"updated_at,omitempty"`
}
