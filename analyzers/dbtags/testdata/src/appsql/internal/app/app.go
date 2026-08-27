// Eval for GID-125: the composition root opens the database.
package app

import "database/sql"

// App wires the service together.
type App struct {
	db *sql.DB
}
