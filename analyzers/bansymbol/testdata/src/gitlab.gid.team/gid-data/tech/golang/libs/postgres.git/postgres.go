// Stub gitlab.gid.team/gid-data/tech/golang/libs/postgres.git для eval.
// Импортируется как gdpostgres.
package postgres

// Conn — заглушка соединения.
type Conn struct{}

// TQuery — запрещённый дженерик-символ (GID-217 / repo.md).
func TQuery[T any](conn *Conn, query string) (T, error) {
	var zero T
	return zero, nil
}

// Select — разрешённый прямой метод conn.
func Select(conn *Conn, dest any, query string) error { return nil }

// NamedStruct — разрешённый прямой метод conn.
func NamedStruct(conn *Conn, arg any) (string, error) { return "", nil }
