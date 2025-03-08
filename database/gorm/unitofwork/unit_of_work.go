package gorm

import (
	"context"
	"database/sql"

	gdk "github.com/loongkirin/gdk/database/gorm"
	uow "github.com/loongkirin/gdk/database/unitofwork"
	"gorm.io/gorm"
)

type unitOfWork struct {
	db           *gorm.DB
	repositories map[string]uow.RepositoryFactory
}

func NewUnitOfWork(dbContext gdk.DbContext) uow.UnitOfWork {
	db := dbContext.GetMasterDb()
	if db == nil {
		return nil
	}
	return &unitOfWork{
		db:           db,
		repositories: make(map[string]uow.RepositoryFactory),
	}
}

func (u *unitOfWork) Register(name string, factory uow.RepositoryFactory) error {
	if _, ok := u.repositories[name]; ok {
		return uow.ErrRepositoryAlreadyRegistered
	}

	u.repositories[name] = factory
	return nil
}

func (u *unitOfWork) Remove(name string) error {
	if _, ok := u.repositories[name]; !ok {
		return uow.ErrRepositoryNotRegistered
	}

	delete(u.repositories, name)
	return nil
}

func (u *unitOfWork) Has(name string) bool {
	_, ok := u.repositories[name]
	return ok
}

func (u *unitOfWork) Clear() {
	u.repositories = make(map[string]uow.RepositoryFactory)
}

func (u *unitOfWork) Do(ctx context.Context, t uow.Transaction, fn uow.SaveChange, opts ...*sql.TxOptions) error {
	tx, err := t.Begin(opts...)
	if err != nil {
		return err
	}
	defer t.Rollback(tx)
	fn(ctx, tx)
	return t.Commit(tx)
}
