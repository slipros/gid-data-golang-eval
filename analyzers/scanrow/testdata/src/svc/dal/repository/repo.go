// Eval of GID-245: an epgx Select into a one-field struct should use ScanRow.
package repository

import "context"

// Conn mirrors the epgx connection surface: Select maps into a struct/slice,
// ScanRow assigns a single row's columns into a slice of scalar pointers.
type Conn interface {
	Select(ctx context.Context, ptr any, sql string, args ...any) error
	ScanRow(ctx context.Context, scan []any, sql string, args ...any) error
	Exec(ctx context.Context, sql string, args ...any) error
}

// OnlySelect has Select but no ScanRow — not confirmed to be an epgx connection.
type OnlySelect interface {
	Select(ctx context.Context, ptr any, sql string, args ...any) error
}

type oneCol struct {
	ID string `db:"id"`
}

type twoCol struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

type Repo struct {
	conn  Conn
	other OnlySelect
}

// --- Class 1: positive ---

func (r *Repo) badAnon(ctx context.Context) error {
	var out struct {
		MemberID string `db:"member_id"`
	}
	return r.conn.Select(ctx, &out, "sql") // want `GID-245: Select into a single-field struct reads one column`
}

func (r *Repo) badNamed(ctx context.Context) error {
	var out oneCol
	return r.conn.Select(ctx, &out, "sql") // want `GID-245: Select into a single-field struct reads one column`
}

// --- Class 2: negative ---

func (r *Repo) goodMulti(ctx context.Context) error {
	var out twoCol
	return r.conn.Select(ctx, &out, "sql")
}

func (r *Repo) goodSlice(ctx context.Context) error {
	var out []oneCol
	return r.conn.Select(ctx, &out, "sql")
}

func (r *Repo) goodScanRow(ctx context.Context) error {
	var out string
	return r.conn.ScanRow(ctx, []any{&out}, "sql")
}

// --- Class 3: boundary ---

func (r *Repo) boundaryNonEpgx(ctx context.Context) error {
	var out oneCol
	return r.other.Select(ctx, &out, "sql")
}
