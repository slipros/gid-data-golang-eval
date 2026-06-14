// Not applicable: a usecase is an orchestrator, it may depend on several
// services (the rule does not even touch structs from its own package —
// it applies only in /domain/service).
package usecase

import "context"

type SnapshotService interface {
	Snapshot(ctx context.Context, id string) (string, error)
}

type JobService interface {
	Job(ctx context.Context, id string) (string, error)
}

type helper struct{}

type Upload struct {
	snapshots SnapshotService
	jobs      JobService
	h         helper
}
