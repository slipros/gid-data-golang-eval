// Package wide — the boundary of the interface side: the package converts the
// type to a wider interface, never to the asserted one.
package wide

// SnapshotPort — the asserted port.
type SnapshotPort interface {
	Snapshot(id string) string
}

// SnapshotAndSchemaPort — the wider port the package actually converts to.
type SnapshotAndSchemaPort interface {
	Snapshot(id string) string
	Schema(id string) string
}

// The conversion below proves that Store implements the wider port, and the
// narrower one follows from it — but naming a contract the reader cannot find
// on the line the diagnostic points at would be worse than staying quiet.
var _ SnapshotPort = (*Store)(nil)

// Store — the implementation.
type Store struct{}

// Snapshot returns the snapshot identifier.
func (s *Store) Snapshot(id string) string { return id }

// Schema returns the schema of the table.
func (s *Store) Schema(id string) string { return id }

// Use takes the wider port.
func Use(port SnapshotAndSchemaPort) string { return port.Snapshot("id") }

// Wire hands the store over as the wider port.
func Wire() string { return Use(&Store{}) }
