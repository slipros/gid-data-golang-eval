// Eval GID-133: private helpers in _test.go are excluded. A shared test-builder
// is package-level by design (it cannot be an entity method), so it must NOT be
// flagged even though it is a private package-level function in a service layer.
package service

// buildSnapshot is a shared test-builder used by several tests. Under GID-133 a
// private package-level function in a service package would be flagged, but
// _test.go is out of scope — no diagnostic here.
func buildSnapshot(name string) *Snapshot {
	return &Snapshot{name: name}
}

var _ = buildSnapshot
