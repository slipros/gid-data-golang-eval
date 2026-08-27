// Eval for GID-125, positive: the module imports the SQL stack in its
// repository, so its entities are mapped onto columns and the rule judges them.
package entity

// Account — an untagged field of a module that really speaks SQL.
type Account struct {
	ID    string `db:"id"`
	Email string // want `GID-125: field Account\.Email has no mapping tag \(db\)\. Fix: add a tag so entity-to-column mapping is explicit`
}
