package nomsg

import gdpostgres "gitlab.gid.team/gid-data/tech/golang/libs/postgres.git"

// settings.symbols configures a symbol without Msg — the diagnostic uses the
// generic wording "symbol %s.%s is banned by gidbansymbol".

func callSelect(conn *gdpostgres.Conn, dest any) error {
	return gdpostgres.Select(conn, dest, "select 1") // want `GID-217: symbol postgres\.Select is banned by gidbansymbol\. Fix: replace it with the project-approved alternative\.`
}
