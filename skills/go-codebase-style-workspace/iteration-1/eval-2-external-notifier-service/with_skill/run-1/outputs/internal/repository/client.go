// Package repository provides the sqlx data access layer. Repositories own
// ALL database access and translate driver errors into typed sentinel errors.
package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/blkst8/client-service/internal/models"
)

// ErrClientNotFound is returned when no client matches the given id.
var ErrClientNotFound = errors.New("client not found")

// Client is the data access interface for the clients table.
type Client interface {
	Create(ctx context.Context, client models.Client) (uint32, error)
	Get(ctx context.Context, id uint32) (*models.Client, error)
}

// client is the sqlx implementation of Client.
type client struct {
	db *sqlx.DB
}

// NewClientRepository returns a Client repository backed by the given db.
func NewClientRepository(db *sqlx.DB) Client {
	return &client{
		db: db,
	}
}

// Create inserts a new client row and returns the generated id.
func (c *client) Create(ctx context.Context, client models.Client) (uint32, error) {
	query := `INSERT INTO clients (name, email, telegram_id) VALUES (:name, :email, :telegram_id)`

	result, err := c.db.NamedExecContext(ctx, query, &client)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint32(id), nil
}

// Get returns the client with the given id, or ErrClientNotFound.
func (c *client) Get(ctx context.Context, id uint32) (*models.Client, error) {
	query := `SELECT * FROM clients WHERE id = ?`

	var client models.Client
	err := c.db.GetContext(ctx, &client, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrClientNotFound
		}
		return nil, err
	}

	return &client, nil
}
