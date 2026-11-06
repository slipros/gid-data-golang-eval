package gdpostgres

import "github.com/jackc/pgx/v5"

type NamedArgs = pgx.NamedArgs

func Query(query string, args NamedArgs) {}
