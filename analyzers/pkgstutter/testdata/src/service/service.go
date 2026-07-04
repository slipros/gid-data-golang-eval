// Eval for GID-193: the suffix case does not fire on other roles' suffixes —
// package service.
package service

// SnapshotService — a stuttering type: service.SnapshotService.
type SnapshotService struct{} // want `GID-193: SnapshotService repeats the package name service\. Fix: from outside it is service\.Snapshot; drop the "Service" suffix and name the symbol after the entity`

// SnapshotRepository — a consumer-side dependency interface (GID-134): the
// suffix "Repository" is not the package name "service", no stutter.
type SnapshotRepository interface {
	Ping() error
}
