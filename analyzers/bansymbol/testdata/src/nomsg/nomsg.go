package nomsg

import gdpostgres "gitlab.gid.team/gid-data/tech/golang/libs/postgres.git"

// settings.symbols задаёт символ без Msg — диагностика использует общую
// формулировку «символ %s.%s запрещён настройками gidbansymbol».

func callSelect(conn *gdpostgres.Conn, dest any) error {
	return gdpostgres.Select(conn, dest, "select 1") // want `GID-217: символ postgres\.Select запрещён настройками gidbansymbol`
}
