// Eval for GID-193: the suffix case — package repository.
package repository

// SnapshotRepository — a stuttering type: from outside it is
// repository.SnapshotRepository; the struct must be named after the entity.
type SnapshotRepository struct{} // want `GID-193: SnapshotRepository repeats the package name repository\. Fix: from outside it is repository\.Snapshot; drop the "Repository" suffix and name the symbol after the entity`

// Snapshot — named after the entity: clean.
type Snapshot struct{}

// Repository — an exact match of the package name: allowed, reads like time.Time.
type Repository struct{}

// jobRepository — an unexported symbol, not visible from outside: not matched.
type jobRepository struct{}
