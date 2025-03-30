package gorm

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/loongkirin/gdk/database/repository"
	"github.com/loongkirin/gdk/database/unitofwork"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type Product struct {
	id     uuid.UUID
	amount uint32
}

func NewProduct(amount uint32) *Product {
	return &Product{
		id:     uuid.New(),
		amount: amount,
	}
}

type Order struct {
	id        uuid.UUID
	productId uuid.UUID
	amount    uint32
}

func NewOrder(productId uuid.UUID, amount uint32) *Order {
	return &Order{
		id:        uuid.New(),
		productId: productId,
		amount:    amount,
	}
}

type Repository[T any] struct {
	db *gorm.DB
}

func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{db: db}
}

func (r *Repository[T]) Add(data *T) (*T, error) {
	err := r.db.Create(data).Error
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *Repository[T]) Update(data *T) (*T, error) {
	err := r.db.Save(data).Error
	if err != nil {
		return nil, err
	}

	return data, nil
}

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func Test_Transaction_NewTransaction(t *testing.T) {
	dbqq := &gorm.DB{}
	repositories := make(map[string]unitofwork.RepositoryFactory)
	transaction := NewTransaction(dbqq, repositories)
	repositories["ProductRepository"] = func(db unitofwork.DB) unitofwork.UOWRepository {
		return NewRepository[Product](db.(*gorm.DB))
	}
	if factory, ok := repositories["ProductRepository"]; ok {
		dd := factory(dbqq).(repository.Repository[Product])
		dd.Add(context.Background(), NewProduct(123))
	}

	// var dfsdf repository.Repository[Product] = NewRepository[Product](dbqq)
	// dfsdf.Add(NewProduct(456))

	assert.NotNil(t, transaction)
}
