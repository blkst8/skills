// Package repository provides the data access layer of the application.
//
// Every repository is defined as an interface with a private sqlx-backed
// implementation, so upper layers depend on abstractions and stay testable.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"yourproject/internal/models"
)

var (
	// ErrClientNotFound is returned when no client matches the given ID.
	ErrClientNotFound = errors.New("client not found")
	// ErrClientAlreadyExists is returned when a unique constraint
	// (e.g. the email column) is violated.
	ErrClientAlreadyExists = errors.New("client already exists")
)

// Client is the data access interface for the clients table.
type Client interface {
	Create(ctx context.Context, client models.Client) (uint32, error)
	Get(ctx context.Context, id uint32) (*models.Client, error)
	Update(ctx context.Context, client models.Client) error
	Delete(ctx context.Context, id uint32) error
}

// client is the private sqlx implementation of the Client repository.
type client struct {
	db *sqlx.DB
}

// NewClientRepository builds a new client repository backed by db.
func NewClientRepository(db *sqlx.DB) Client {
	return &client{
		db: db,
	}
}

// Create inserts a new client and returns the auto-generated ID.
func (c *client) Create(ctx context.Context, client models.Client) (uint32, error) {
	query := `INSERT INTO clients (name, email) VALUES (:name, :email)`

	result, err := c.db.NamedExecContext(ctx, query, &client)
	if err != nil {
		if isDuplicateKeyError(err) {
			return 0, ErrClientAlreadyExists
		}
		return 0, fmt.Errorf("failed to insert client: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return uint32(id), nil
}

// Get fetches a single client by ID. Returns ErrClientNotFound when the
// client does not exist.
func (c *client) Get(ctx context.Context, id uint32) (*models.Client, error) {
	query := `SELECT id, name, email, created_at, updated_at FROM clients WHERE id = ?`

	var client models.Client
	if err := c.db.GetContext(ctx, &client, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrClientNotFound
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return &client, nil
}

// Update modifies an existing client. Returns ErrClientNotFound when no row
// was affected.
func (c *client) Update(ctx context.Context, client models.Client) error {
	query := `UPDATE clients SET name = :name, email = :email WHERE id = :id`

	result, err := c.db.NamedExecContext(ctx, query, &client)
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrClientAlreadyExists
		}
		return fmt.Errorf("failed to update client: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrClientNotFound
	}

	return nil
}

// Delete removes a client by ID. Returns ErrClientNotFound when no row was
// affected.
func (c *client) Delete(ctx context.Context, id uint32) error {
	query := `DELETE FROM clients WHERE id = ?`

	result, err := c.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrClientNotFound
	}

	return nil
}

// isDuplicateKeyError reports whether err is a MySQL duplicate entry error
// (error code 1062).
func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
