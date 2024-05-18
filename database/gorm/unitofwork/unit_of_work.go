package gorm

import (
	"context"
	"database/sql"
	"errors"

	gdk "github.com/loongkirin/gdk/database/gorm"
	"github.com/loongkirin/gdk/database/unitofwork"
	"gorm.io/gorm"
)

var (
	ErrRepositoryNotRegistered     = errors.New("repository not registered")
	ErrRepositoryAlreadyRegistered = errors.New("repository already registered")
	ErrInvalidRepositoryType       = errors.New("invalid repository type")
)

type UnitOfWork struct {
	db           *gorm.DB
	repositories map[string]unitofwork.RepositoryFactory
}

func NewUnitOfWork(dbContext gdk.DbContext) unitofwork.UnitOfWork {
	db := dbContext.GetDb()
	if db == nil {
		return nil
	}
	return &UnitOfWork{
		db:           db,
		repositories: make(map[string]unitofwork.RepositoryFactory),
	}
}

func (u *UnitOfWork) Register(name string, factory unitofwork.RepositoryFactory) error {
	if _, ok := u.repositories[name]; ok {
		return ErrRepositoryAlreadyRegistered
	}

	u.repositories[name] = factory
	return nil
}

func (u *UnitOfWork) Remove(name string) error {
	if _, ok := u.repositories[name]; !ok {
		return ErrRepositoryNotRegistered
	}

	delete(u.repositories, name)
	return nil
}

func (u *UnitOfWork) Has(name string) bool {
	_, ok := u.repositories[name]
	return ok
}

func (u *UnitOfWork) Clear() {
	u.repositories = make(map[string]unitofwork.RepositoryFactory)
}

func (u *UnitOfWork) Do(ctx context.Context, t unitofwork.Transaction, fn func(ctx context.Context, tx unitofwork.TxHandler) error, opts ...*sql.TxOptions) error {
	tx, err := t.Begin(opts...)
	if err != nil {
		return err
	}
	defer t.Rollback(tx)
	fn(ctx, tx)
	return t.Commit(tx)
}
