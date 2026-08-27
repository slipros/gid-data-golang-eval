// Eval for GID-125, boundary: the SQL import of the module sits in the
// composition root, not in the dal — the verdict is per module, so the entity
// is judged all the same.
package entity

// Session — untagged, in a module holding a database connection.
type Session struct {
	Token string // want `GID-125: field Session\.Token has no mapping tag \(db\)`
}
