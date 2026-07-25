// Positive: a SQL pool reached from the service layer, past the repository.
package service

import "github.com/jackc/pgx/v5/pgxpool" // want `GID-249: package "svc/domain/service" reaches a data store directly — driver "github.com/jackc/pgx/v5/pgxpool" belongs to the repository layer\. Fix: move the storage code to /dal/repository`

type Service struct {
	pool *pgxpool.Pool
}
