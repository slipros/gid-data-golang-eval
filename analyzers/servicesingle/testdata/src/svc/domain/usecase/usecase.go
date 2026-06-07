// Неприменимость: usecase — оркестратор, ему можно зависеть от
// нескольких сервисов (даже структур из своего пакета правило не касается —
// оно действует только в /domain/service).
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
