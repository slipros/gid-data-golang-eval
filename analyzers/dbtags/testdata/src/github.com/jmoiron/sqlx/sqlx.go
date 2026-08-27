// Stub of github.com/jmoiron/sqlx for eval.
package sqlx

type DB struct{}

func Connect(driver, dsn string) (*DB, error) { return &DB{}, nil }
