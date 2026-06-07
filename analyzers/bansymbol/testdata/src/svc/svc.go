package svc

import (
	gdpostgres "gitlab.gid.team/gid-data/tech/golang/libs/postgres.git"

	"example.com/otherdb"
)

// --- Класс 1: позитивный (нарушение ловится) ---

// Прямой вызов запрещённого gdpostgres.TQuery.
func callTQuery(conn *gdpostgres.Conn) (int, error) {
	return gdpostgres.TQuery[int](conn, "select 1") // want `GID-217: используй прямые методы conn: Select, ScanRow, NamedStruct, Transaction \(repo\.md\)`
}

// Generic-инстанциация с явным типовым аргументом — резолвится так же.
func callTQueryString(conn *gdpostgres.Conn) (string, error) {
	return gdpostgres.TQuery[string](conn, "select name") // want `GID-217: используй прямые методы conn: Select, ScanRow, NamedStruct, Transaction \(repo\.md\)`
}

// --- Класс 2: негативный (чистый код проходит) ---

// Разрешённые прямые методы conn.
func callSelect(conn *gdpostgres.Conn, dest any) error {
	return gdpostgres.Select(conn, dest, "select 1")
}

func callNamedStruct(conn *gdpostgres.Conn, arg any) (string, error) {
	return gdpostgres.NamedStruct(conn, arg)
}

// TQuery из ДРУГОГО пакета с тем же именем — не флагается.
func callOtherTQuery() (int, error) {
	return otherdb.TQuery[int]("select 1")
}
