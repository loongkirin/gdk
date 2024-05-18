package gorm

import (
	"database/sql"

	"github.com/loongkirin/gdk/database/unitofwork"
	uow "github.com/loongkirin/gdk/database/unitofwork"
	"gorm.io/gorm"
)

type transaction struct {
	db           *gorm.DB
	repositories map[string]unitofwork.RepositoryFactory
	Error        error
}

func NewTransaction(db *gorm.DB, repositories map[string]unitofwork.RepositoryFactory) uow.Transaction {
	return &transaction{
		db:           db,
		repositories: repositories,
		Error:        nil,
	}
}

func (t *transaction) Begin(opts ...*sql.TxOptions) (uow.TxHandler, error) {
	tx := t.db.Begin(opts...)
	t.db = tx
	t.Error = tx.Error
	return t, tx.Error
}

func (t *transaction) Rollback(tx uow.TxHandler) error {
	db := tx.(transaction).db
	db = db.Rollback()
	t.db = db
	t.Error = db.Error
	return db.Error
}

func (t *transaction) Commit(tx uow.TxHandler) error {
	db := tx.(transaction).db
	db = db.Commit()
	t.db = db
	t.Error = db.Error
	return db.Error
}
