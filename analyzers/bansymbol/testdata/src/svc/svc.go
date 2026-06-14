package svc

import (
	gdpostgres "gitlab.gid.team/gid-data/tech/golang/libs/postgres.git"

	"example.com/otherdb"
)

// --- Class 1: positive (the violation is caught) ---

// A direct call of the banned gdpostgres.TQuery.
func callTQuery(conn *gdpostgres.Conn) (int, error) {
	return gdpostgres.TQuery[int](conn, "select 1") // want `GID-217: gdpostgres\.TQuery is banned\. Fix: use conn methods directly: Select, ScanRow, NamedStruct or Transaction \(repo\.md\)`
}

// A generic instantiation with an explicit type argument — resolves the same way.
func callTQueryString(conn *gdpostgres.Conn) (string, error) {
	return gdpostgres.TQuery[string](conn, "select name") // want `GID-217: gdpostgres\.TQuery is banned\. Fix: use conn methods directly: Select, ScanRow, NamedStruct or Transaction \(repo\.md\)`
}

// --- Class 2: negative (clean code passes) ---

// Allowed direct conn methods.
func callSelect(conn *gdpostgres.Conn, dest any) error {
	return gdpostgres.Select(conn, dest, "select 1")
}

func callNamedStruct(conn *gdpostgres.Conn, arg any) (string, error) {
	return gdpostgres.NamedStruct(conn, arg)
}

// TQuery from a DIFFERENT package with the same name — not flagged.
func callOtherTQuery() (int, error) {
	return otherdb.TQuery[int]("select 1")
}
