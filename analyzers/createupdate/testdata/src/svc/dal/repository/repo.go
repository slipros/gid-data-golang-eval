// Eval for GID-112 (repo).
package repository

import "context"

type Snapshot struct{ ID string }

type Job struct{}

// --- Positive: Create returns data ---

func (j *Job) CreateJob(ctx context.Context, name string) (Snapshot, error) { // want `GID-112: method "CreateJob" creates/updates state and must return only error`
	return Snapshot{}, nil
}

// Edge case: Update without error at all.
func (j *Job) UpdateJobStatus(ctx context.Context, status string) Snapshot { // want `GID-112: method "UpdateJobStatus" creates/updates state and must return only error`
	return Snapshot{}
}

// --- Negative: only error ---

func (j *Job) CreateJobs(ctx context.Context, names []string) error {
	return nil
}

func (j *Job) UpdateJob(ctx context.Context, id string) error {
	return nil
}

// Edge case: Created is not the verb Create.
func (j *Job) CreatedJobs(ctx context.Context) ([]Snapshot, error) {
	return nil, nil
}

// Not applicable: fetching is not creating/updating.
func (j *Job) Job(ctx context.Context, id string) (Snapshot, error) {
	return Snapshot{}, nil
}
