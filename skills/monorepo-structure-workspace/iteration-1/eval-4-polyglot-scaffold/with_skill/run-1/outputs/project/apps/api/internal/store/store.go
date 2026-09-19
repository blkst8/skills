// Package store owns all database access (pgx). Nothing outside apps/api
// touches this — the frontend knows the world through the OpenAPI contract
// in packages/api-schema, never through database internals.
package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// User is the persisted user record. Its JSON tags are the wire format and
// must stay in sync with the User schema in packages/api-schema/openapi.yaml.
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.pool.Query(ctx, `select id, email, name from users order by created_at desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *Store) CreateUser(ctx context.Context, email, name string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`insert into users (email, name) values ($1, $2) returning id, email, name`,
		email, name,
	).Scan(&u.ID, &u.Email, &u.Name)
	if err != nil {
		return User{}, err
	}
	return u, nil
}
