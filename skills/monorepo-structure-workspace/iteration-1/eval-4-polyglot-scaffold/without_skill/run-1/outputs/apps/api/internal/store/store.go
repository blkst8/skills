package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/blkst8/sideproject/apps/api/internal/apitypes"
)

// Store wraps all database access. Handlers depend on this interface, not on
// pgx directly, which keeps them testable with a fake store.
type Store interface {
	ListItems(ctx context.Context) ([]apitypes.Item, error)
}

type pgStore struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) Store {
	return &pgStore{pool: pool}
}

func (s *pgStore) ListItems(ctx context.Context) ([]apitypes.Item, error) {
	rows, err := s.pool.Query(ctx, `select id, title, created_at from items order by created_at desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]apitypes.Item, 0)
	for rows.Next() {
		var it apitypes.Item
		if err := rows.Scan(&it.ID, &it.Title, &it.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}
