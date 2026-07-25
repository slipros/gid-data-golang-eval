// Negative: the composition root opens the pool and injects it into repositories.
package api

import "github.com/jackc/pgx/v5/pgxpool"

type Application struct {
	pool *pgxpool.Pool
}
