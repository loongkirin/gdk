package unitofwork

import (
	"database/sql"
)

type TxHandler any

type Transaction interface {
	Begin(opts ...*sql.TxOptions) (TxHandler, error)
	Rollback(tx TxHandler) error
	Commit(tx TxHandler) error
}
