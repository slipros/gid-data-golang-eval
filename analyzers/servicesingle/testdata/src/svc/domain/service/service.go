// Eval для GID-148 (service-single).
package service

import "context"

// SnapshotRepository — интерфейс зависимости, определён рядом с потребителем.
type SnapshotRepository interface {
	Snapshot(ctx context.Context, id string) (string, error)
}

// SnapshotOptions — настройки сервиса.
type SnapshotOptions struct {
	MaxSize int
}

// Snapshot — сервис, посвящённый сущности Snapshot.
type Snapshot struct {
	repo SnapshotRepository
	opts SnapshotOptions
}

// Job — второй сервис в том же пакете.
type Job struct {
	repo SnapshotRepository
}

// --- Позитив: сервис зависит от другого сервиса ---

type Upload struct {
	snapshots *Snapshot // want `GID-148: service "Upload" depends on service "Snapshot"\. Fix: a service serves one entity, orchestrate multiple services in usecase`
	jobs      Job       // want `GID-148: service "Upload" depends on service "Job"`
}

// --- Негатив: зависимости-интерфейсы и Options — норма (см. Snapshot, Job) ---
