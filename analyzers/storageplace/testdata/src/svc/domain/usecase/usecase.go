// Positive: database/sql used for a connection, not just NULL types.
package usecase

import "database/sql" // want `GID-249: package "svc/domain/usecase" reaches a data store directly — driver "database/sql" belongs to the repository layer`

type Usecase struct {
	db *sql.DB
}
