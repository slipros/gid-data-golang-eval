// Eval for GID-194 (constscope): a service package — the rule's ordinary scope.
package service

// --- Positive: an exported constant outside model/entity ---

const DefaultPageSize = 25 // want `GID-194: exported constant "DefaultPageSize" is declared outside model/entity\. Fix: keep shared constants in /domain/model or /dal/entity, and declare local ones where they are used`

// --- Positive: a constant used by only one method ---

const snapshotPrefix = "snap-" // want `GID-194: constant "snapshotPrefix" is used only in "Snapshot\.Render"\. Fix: declare it inside that function`

// --- Negative: a constant shared by two methods — package-level is legal ---

const snapshotTable = "snapshots"

type Snapshot struct{}

func (s *Snapshot) Render() string {
	return snapshotPrefix + snapshotTable
}

func (s *Snapshot) Table() string {
	return snapshotTable
}

// --- Negative: a constant declared inside a function — the target state ---

func (s *Snapshot) Endpoint() string {
	const endpoint = "/v1/snapshots"
	return endpoint
}
