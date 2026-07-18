// Eval GID-168: boundary — a "domain" segment nested under another layer
// (server), not anchored to the module root, is not the domain layer.
// Without the HasLayer fix, pathseg.Contains would have falsely matched the
// "domain" segment here and flagged the db tag below.
package model

// Boundary class: a db tag in server/domain/model — "domain" is not the
// leading (root-anchored) layer segment here, so GID-168 does not apply.

type Snapshot struct {
	ID string `db:"id"`
}
