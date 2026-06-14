package custom

import "example.com/otherdb"

// settings.symbols configures the custom symbol otherdb.TQuery with a custom Msg —
// it is flagged with exactly that message (and the default gdpostgres.TQuery is
// not configured in this run).

func callBanned() (int, error) {
	return otherdb.TQuery[int]("select 1") // want `GID-217: otherdb\.TQuery is banned by the project`
}
