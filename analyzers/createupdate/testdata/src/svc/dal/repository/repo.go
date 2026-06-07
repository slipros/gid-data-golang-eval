// Eval для GID-112 (repo).
package repository

import "context"

type Snapshot struct{ ID string }

type Job struct{}

// --- Позитив: Create возвращает данные ---

func (j *Job) CreateJob(ctx context.Context, name string) (Snapshot, error) { // want `GID-112: method "CreateJob" creates/updates state and must return only error`
	return Snapshot{}, nil
}

// Граничный кейс: Update без error вовсе.
func (j *Job) UpdateJobStatus(ctx context.Context, status string) Snapshot { // want `GID-112: method "UpdateJobStatus" creates/updates state and must return only error`
	return Snapshot{}
}

// --- Негатив: только error ---

func (j *Job) CreateJobs(ctx context.Context, names []string) error {
	return nil
}

func (j *Job) UpdateJob(ctx context.Context, id string) error {
	return nil
}

// Граничный кейс: Created — не глагол Create.
func (j *Job) CreatedJobs(ctx context.Context) ([]Snapshot, error) {
	return nil, nil
}

// Неприменимость: получение — не создание/обновление.
func (j *Job) Job(ctx context.Context, id string) (Snapshot, error) {
	return Snapshot{}, nil
}
