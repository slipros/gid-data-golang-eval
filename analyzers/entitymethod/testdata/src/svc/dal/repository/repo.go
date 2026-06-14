// Eval for GID-114 (repo): the root package /dal/repository is in scope.
package repository

import "context"

type Snapshot struct{ ID string }

type Job struct{}

// --- Class 1: positive (the violation is caught) ---

// The List prefix is forbidden.
func (j *Job) ListJobs(ctx context.Context) ([]Snapshot, error) { // want `GID-114: drop the List prefix\. Fix: use the plural Jobs instead of ListJobs`
	return nil, nil
}

// The exact ByID suffix is forbidden.
func (j *Job) JobByID(ctx context.Context, id string) (Snapshot, error) { // want `GID-114: drop the ByID suffix\. Fix: use Job\(ctx, id\) instead of JobByID`
	return Snapshot{}, nil
}

// The method name does not contain the Job entity name.
func (j *Job) Fetch(ctx context.Context) (Snapshot, error) { // want `GID-114: method name "Fetch" must contain the entity name "Job"`
	return Snapshot{}, nil
}

// FP zone: a verb method without an entity — rarely legitimate, but caught; muted via exclude/nolint.
func (j *Job) Close() error { // want `GID-114: method name "Close" must contain the entity name "Job"`
	return nil
}

// --- Class 2: negative (clean code passes) ---

func (j *Job) Job(ctx context.Context, id string) (Snapshot, error) {
	return Snapshot{}, nil
}

func (j *Job) Jobs(ctx context.Context) ([]Snapshot, error) {
	return nil, nil
}

func (j *Job) CreateJob(ctx context.Context, name string) error {
	return nil
}

func (j *Job) DeleteJob(ctx context.Context, id string) error {
	return nil
}

// --- Class 3: edge ---

// ByStageID is a query refinement, not the ByID suffix; the name contains Job. Allowed.
func (j *Job) JobsByStageID(ctx context.Context, stageID string) ([]Snapshot, error) {
	return nil, nil
}

// An unexported method is not matched.
func (j *Job) listJobsInternal(ctx context.Context) ([]Snapshot, error) {
	return nil, nil
}

// Word boundary: Listen is not the List prefix (the next rune is lowercase).
func (j *Job) ListenJobEvents(ctx context.Context) error {
	return nil
}

// FP zone: the verb method Ping; in production muted via //nolint:gidentitymethod
// (analysistest does not process nolint — filtering happens on the golangci-lint
// side, so the diagnostic is expected here).
func (j *Job) Ping(ctx context.Context) error { // want `GID-114: method name "Ping" must contain the entity name "Job"`
	return nil
}
