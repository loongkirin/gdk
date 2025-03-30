package repository

import (
	"context"

	"github.com/loongkirin/gdk/database/query"
)

type Repository[T any] interface {
	Migrate(ctx context.Context, data *T) error
	QueryById(ctx context.Context, id string) (*T, error)
	Query(ctx context.Context, query *query.DbQuery) ([]T, error)
	Add(ctx context.Context, data *T) (*T, error)
	Update(ctx context.Context, data *T) (*T, error)
	Delete(ctx context.Context, data *T) (bool, error)
}
