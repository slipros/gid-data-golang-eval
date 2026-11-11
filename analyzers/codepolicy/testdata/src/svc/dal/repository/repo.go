package repository

import gdpostgres "gitlab.gid.team/gid-data/tech/golang/libs/postgres.git"

type Datamart struct{ ID int }

type DatamartTitles = Datamart // want `GID-282:`

// Boundary: a defined type is a distinct contract.
type DatamartCopy Datamart

const positional = `SELECT id FROM datamarts WHERE id = $1` // want `GID-284:`
const multiline = `
 SELECT id
 FROM datamarts
 WHERE id = @id
`
const shortSQL = "SELECT 1"
const literalDollar = `SELECT '$1', "$2", $$ $3 $$, $tag$ $4 $tag$ /* $5 */ -- $6`
const longSQL = "SELECT identifier, description, creation_timestamp, update_timestamp FROM datamarts WHERE enabled = true" // want `GID-283:`

func inline() {
	gdpostgres.Query("SELECT id FROM datamarts WHERE id = @id", gdpostgres.NamedArgs{"id": 1}) // want `GID-285:` `GID-297:`
}

func explicit() {
	args := gdpostgres.NamedArgs{"id": 1}
	gdpostgres.Query(multiline, args)
}

type OtherNamedArgs map[string]any

func unrelated(args OtherNamedArgs) {}

func negative() {
	unrelated(OtherNamedArgs{"id": 1})
}

// want +1 `GID-283:`
const fakeMultiline = `
SELECT identifier, description, creation_timestamp, update_timestamp FROM datamarts WHERE enabled = true
`

func inlineSQLPlaceholder() {
	args := gdpostgres.NamedArgs{"id": 1}
	gdpostgres.Query("SELECT id FROM datamarts WHERE id = $1", args) // want `GID-284:` `GID-297:`
}
