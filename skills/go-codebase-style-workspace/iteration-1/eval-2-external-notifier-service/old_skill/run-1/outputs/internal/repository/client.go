// Package repository provides the data access layer.
//
// Every repository is interface-driven: the interface, a private
// implementation struct, and a public constructor returning the interface.
// Sentinel errors (ErrClientNotFound, ...) are defined here so upper layers
// can map them with errors.Is.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/blkst8/client-service/internal/models"
)

// ErrClientNotFound is returned when no client matches the given identifier.
var ErrClientNotFound = errors.New("client not found")

// Client defines the data access operations for clients.
type Client interface {
	// Create inserts a new client and populates client.ID with the
	// generated primary key.
	Create(ctx context.Context, client *models.Client) error
	// Get returns the client with the given ID or ErrClientNotFound.
	Get(ctx context.Context, id uint32) (*models.Client, error)
}

// client is the MySQL implementation of the Client repository.
type client struct {
	db *sqlx.DB
}

// NewClientRepository returns a MySQL-backed Client repository.
func NewClientRepository(db *sqlx.DB) Client {
	return &client{
		db: db,
	}
}

// Create inserts the client with a named query and writes the generated ID
// back into client.
func (c *client) Create(ctx context.Context, client *models.Client) error {
	query := `INSERT INTO clients (name, email, telegram_id) VALUES (:name, :email, :telegram_id)`

	result, err := c.db.NamedExecContext(ctx, query, client)
	if err != nil {
		return fmt.Errorf("failed to insert client: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to read last insert id: %w", err)
	}
	client.ID = uint32(id)

	return nil
}

// Get fetches a single client by ID and maps sql.ErrNoRows to
// ErrClientNotFound.
func (c *client) Get(ctx context.Context, id uint32) (*models.Client, error) {
	query := `SELECT id, name, email, telegram_id, created_at, updated_at FROM clients WHERE id = ?`

	var client models.Client
	if err := c.db.GetContext(ctx, &client, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrClientNotFound
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return &client, nil
}
