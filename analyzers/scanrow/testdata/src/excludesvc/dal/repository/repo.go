// Eval of GID-245 settings.exclude: an exempted method is not flagged even
// though it matches the anti-pattern.
package repository

import "context"

type Conn interface {
	Select(ctx context.Context, ptr any, sql string, args ...any) error
	ScanRow(ctx context.Context, scan []any, sql string, args ...any) error
}

type oneCol struct {
	ID string `db:"id"`
}

type Repo struct {
	conn Conn
}

// excludedMethod matches the anti-pattern but is listed in settings.exclude
// ("Repo.excludedMethod") — no diagnostic.
func (r *Repo) excludedMethod(ctx context.Context) error {
	var out oneCol
	return r.conn.Select(ctx, &out, "sql")
}
