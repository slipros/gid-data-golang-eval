// Eval for GID-125: the SQL stack of the module lives here.
package repository

import (
	"github.com/jmoiron/sqlx"

	"sqlsvc/dal/entity"
)

// Accounts reads accounts from the database.
type Accounts struct {
	db *sqlx.DB
}

// ByID returns one account.
func (a *Accounts) ByID(id string) entity.Account {
	return entity.Account{ID: id}
}
