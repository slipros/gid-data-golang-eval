// Eval для GID-114 (repo): корневой пакет /dal/repository — в scope.
package repository

import "context"

type Snapshot struct{ ID string }

type Job struct{}

// --- Класс 1: позитив (нарушение ловится) ---

// Префикс List запрещён.
func (j *Job) ListJobs(ctx context.Context) ([]Snapshot, error) { // want `GID-114: drop the List prefix\. Fix: use the plural Jobs instead of ListJobs`
	return nil, nil
}

// Точный суффикс ByID запрещён.
func (j *Job) JobByID(ctx context.Context, id string) (Snapshot, error) { // want `GID-114: drop the ByID suffix\. Fix: use Job\(ctx, id\) instead of JobByID`
	return Snapshot{}, nil
}

// Имя метода не содержит имя сущности Job.
func (j *Job) Fetch(ctx context.Context) (Snapshot, error) { // want `GID-114: method name "Fetch" must contain the entity name "Job"`
	return Snapshot{}, nil
}

// FP-зона: метод-глагол без сущности — легитимен редко, но ловится; гасится exclude/nolint.
func (j *Job) Close() error { // want `GID-114: method name "Close" must contain the entity name "Job"`
	return nil
}

// --- Класс 2: негатив (чистый код проходит) ---

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

// --- Класс 3: граничный ---

// ByStageID — уточнение выборки, не суффикс ByID; имя содержит Job. Разрешено.
func (j *Job) JobsByStageID(ctx context.Context, stageID string) ([]Snapshot, error) {
	return nil, nil
}

// Неэкспортируемый метод не матчится.
func (j *Job) listJobsInternal(ctx context.Context) ([]Snapshot, error) {
	return nil, nil
}

// Граница слова: Listen — не префикс List (следующая руна строчная).
func (j *Job) ListenJobEvents(ctx context.Context) error {
	return nil
}

// FP-зона: метод-глагол Ping; в проде гасится //nolint:gidentitymethod
// (analysistest не обрабатывает nolint — фильтрация на стороне golangci-lint,
// поэтому здесь диагностика ожидаема).
func (j *Job) Ping(ctx context.Context) error { // want `GID-114: method name "Ping" must contain the entity name "Job"`
	return nil
}
