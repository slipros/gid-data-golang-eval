// Package lib — a library package: it exports an implementation and converts
// nothing itself, so the assertion is the only place the contract is checked.
package lib

import (
	"svc/service"
	"svc/usecase"
)

// The consumer lives outside this module. Without the assertion nothing here
// breaks when the port grows a method — the rule stays silent.
var _ service.DatasetSnapshotRepository = (*Snapshot)(nil)

// An interface asserted against another interface is not the wiring shape the
// rule is about, and is left alone even when the package converts it.
var _ usecase.LatestPageSnapshot = (SnapshotPort)(nil)

// An assertion of the empty interface states no contract to begin with.
var _ any = (*Snapshot)(nil)

// SnapshotPort — the port of the library.
type SnapshotPort interface {
	Snapshot(id string) string
}

// Snapshot — the exported implementation of the port.
type Snapshot struct{}

// Snapshot returns the snapshot identifier.
func (s *Snapshot) Snapshot(id string) string { return id }

// Wire hands the port over as the usecase interface — the interface-to-interface
// assertion above is still not judged.
func Wire(port SnapshotPort) string {
	var page usecase.LatestPageSnapshot = port

	return page.Snapshot("id")
}
