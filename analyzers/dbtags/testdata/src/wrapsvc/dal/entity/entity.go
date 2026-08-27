// Eval for GID-125 with settings.sql-imports: the module reaches the database
// through an in-house wrapper, so the default stack does not see it — naming
// the wrapper in the settings turns the rule back on.
package entity

// Client — untagged, judged only because the wrapper is on the list.
type Client struct {
	ID   string `db:"id"`
	Name string // want `GID-125: field Client\.Name has no mapping tag \(db\)`
}
