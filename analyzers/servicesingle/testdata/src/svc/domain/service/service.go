// Eval for GID-148 (service-single).
package service

import "context"

// SnapshotRepository — a dependency interface, defined next to its consumer.
type SnapshotRepository interface {
	Snapshot(ctx context.Context, id string) (string, error)
}

// SnapshotOptions — the service settings.
type SnapshotOptions struct {
	MaxSize int
}

// Snapshot — the service devoted to the Snapshot entity.
type Snapshot struct {
	repo SnapshotRepository
	opts SnapshotOptions
}

// Job — a second service in the same package.
type Job struct {
	repo SnapshotRepository
}

// --- Positive: a service depends on another service ---

type Upload struct {
	snapshots *Snapshot // want `GID-148: service "Upload" depends on service "Snapshot"\. Fix: a service serves one entity, orchestrate multiple services in usecase`
	jobs      Job       // want `GID-148: service "Upload" depends on service "Job"`
}

// --- Negative: interface dependencies and Options are the norm (see Snapshot, Job) ---

func (s *Job) prepare() {
	preparer := &Snapshot{repo: s.repo} // want `GID-148: service "Job" constructs service "Snapshot" inside a method\. Fix:`
	_ = preparer
	_ = Snapshot{}          // want `GID-148: service "Job" constructs service "Snapshot" inside a method`
	_ = new(Snapshot)       // want `GID-148: service "Job" constructs service "Snapshot" inside a method`
	_ = NewSnapshot(s.repo) // want `GID-148: service "Job" constructs service "Snapshot" inside a method`
	_ = &SnapshotOptions{}
	_ = &Job{repo: s.repo}
}

func NewSnapshot(repo SnapshotRepository) *Snapshot {
	return &Snapshot{repo: repo}
}

func NewJob(repo SnapshotRepository) *Job {
	_ = NewSnapshot(repo)
	return &Job{repo: repo}
}

type SnapshotAlias = Snapshot

func (s Job) byValue() {
	_ = &SnapshotAlias{} // want `GID-148: service "Job" constructs service "Snapshot" inside a method`
}
