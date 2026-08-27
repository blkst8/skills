package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"yourproject/internal/models"
)

var ErrClientNotFound = errors.New("client not found")

// Interface definition
type Client interface {
	Create(ctx context.Context, client *models.Client) error
	Get(ctx context.Context, id uint32) (*models.Client, error)
	Update(ctx context.Context, client models.Client) error
	Delete(ctx context.Context, id uint32) error
}

// Implementation struct (private)
type client struct {
	db *sqlx.DB
}

// Constructor
func NewClientRepository(db *sqlx.DB) Client {
	return &client{
		db: db,
	}
}

func (c *client) Create(ctx context.Context, client *models.Client) error {
	query := `INSERT INTO clients (name, email) VALUES (:name, :email)`
	res, err := c.db.NamedExecContext(ctx, query, client)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	client.ID = uint32(id)

	return nil
}

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

func (c *client) Update(ctx context.Context, client models.Client) error {
	query := `UPDATE clients SET name = :name, email = :email WHERE id = :id`
	res, err := c.db.NamedExecContext(ctx, query, &client)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrClientNotFound
	}

	return nil
}

func (c *client) Delete(ctx context.Context, id uint32) error {
	query := `DELETE FROM clients WHERE id = ?`
	res, err := c.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrClientNotFound
	}

	return nil
}
