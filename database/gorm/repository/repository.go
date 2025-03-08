package repository

import (
	"context"
	"errors"

	"github.com/loongkirin/gdk/database/query"
	"gorm.io/gorm"
)

type Repository[T any] struct {
	db *gorm.DB
}

func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{
		db: db,
	}
}

func (r *Repository[T]) Migrate(ctx context.Context, data *T) error {
	return r.db.WithContext(ctx).AutoMigrate(data)
}

func (r *Repository[T]) QueryById(ctx context.Context, id string) (*T, error) {
	data := new(T)
	err := r.db.WithContext(ctx).Where("id=?", id).First(&data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return data, nil
}

func (r *Repository[T]) Query(ctx context.Context, query *query.DbQuery) ([]T, error) {
	datas := []T{}
	whereClaues, values, order := query.GetWhereClause()
	offset := (query.PageNumber - 1) * query.PageSize
	err := r.db.WithContext(ctx).Where(whereClaues, values...).Order(order).Offset(offset).Limit(query.PageSize + 1).Find(&datas).Error
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (r *Repository[T]) Add(ctx context.Context, data *T) (*T, error) {
	err := r.db.WithContext(ctx).Create(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Repository[T]) Update(ctx context.Context, data *T) (*T, error) {
	err := r.db.WithContext(ctx).Save(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Repository[T]) Delete(ctx context.Context, data *T) (bool, error) {
	err := r.db.WithContext(ctx).Delete(data).Error
	if err != nil {
		return false, err
	}
	return true, nil
}
