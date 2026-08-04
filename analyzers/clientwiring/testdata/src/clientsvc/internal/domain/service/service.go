// Eval of GID-255 non-applicability: another layer entirely. A function-only
// package outside /client/** is none of this rule's business (GID-133 governs
// package functions in the service layer).
package service

// Snapshot is a domain value.
type Snapshot struct {
	ID string
}

// NewSnapshot builds a snapshot.
func NewSnapshot(id string) Snapshot {
	return Snapshot{ID: id}
}
