// Stub of gitlab.gid.team/gid-data/tech/golang/libs/postgres.git for eval.
// Imported as gdpostgres.
package postgres

// Conn — a connection stub.
type Conn struct{}

// TQuery — the banned generic symbol (GID-217 / repo.md).
func TQuery[T any](conn *Conn, query string) (T, error) {
	var zero T
	return zero, nil
}

// Select — an allowed direct conn method.
func Select(conn *Conn, dest any, query string) error { return nil }

// NamedStruct — an allowed direct conn method.
func NamedStruct(conn *Conn, arg any) (string, error) { return "", nil }
