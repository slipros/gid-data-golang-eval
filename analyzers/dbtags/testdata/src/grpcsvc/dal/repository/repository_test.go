// Eval for GID-125: a SQL import in a _test.go file does not make the module
// speak SQL — a fixture may reach for the driver in a service that stores
// nothing.
package repository

import "github.com/jmoiron/sqlx"

var testDB *sqlx.DB
