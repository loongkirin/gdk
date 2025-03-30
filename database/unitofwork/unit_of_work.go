package unitofwork

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrRepositoryNotRegistered     = errors.New("repository not registered")
	ErrRepositoryAlreadyRegistered = errors.New("repository already registered")
	ErrInvalidRepositoryType       = errors.New("invalid repository type")
)

type DB any
type UOWRepository any
type RepositoryFactory func(db DB) UOWRepository
type SaveChange func(ctx context.Context, tx TxHandler) error

type UnitOfWork interface {
	Register(name string, factory RepositoryFactory) error
	Remove(name string) error
	Has(name string) bool
	Clear()
	Do(ctx context.Context, t Transaction, fn SaveChange, opts ...*sql.TxOptions) error
}
