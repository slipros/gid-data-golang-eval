// Stub of an in-house Postgres wrapper hiding the driver from the service.
package pgstore

type Store struct{}

func Open(dsn string) (*Store, error) { return &Store{}, nil }
