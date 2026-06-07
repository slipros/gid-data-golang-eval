// Eval для GID-175: проверка 4 (tx-метод на репозитории).
package repository

import "context"

type JobRepository struct{}

// --- Проверка 4 (позитив): tx-метод на репозитории (имя любое) ---

func (r *JobRepository) InTx(ctx context.Context, fn func(ctx context.Context) error) error { // want `GID-175: a repository/service must not wrap a transaction in a method`
	return fn(ctx)
}

// --- Граничный кейс: метод с похожей сигнатурой, callback возвращает не error — не флагуем ---

func (r *JobRepository) NotInTx(ctx context.Context, fn func(ctx context.Context) (int, error)) error {
	return nil
}

// --- Негатив: обычный метод репозитория ---

func (r *JobRepository) Job(ctx context.Context, id string) error {
	return nil
}
