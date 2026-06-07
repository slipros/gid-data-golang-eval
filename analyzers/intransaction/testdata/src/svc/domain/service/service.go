// Eval для GID-175: проверки 3 (анонимная сигнатура) и 4 (tx-метод на сервисе).
package service

import (
	"context"

	"svc/domain/model"
)

// --- Негатив: поле именованного типа model.InTransactionFunc — ок ---

type JobService struct {
	tx model.InTransactionFunc
}

func NewJobService(tx model.InTransactionFunc) *JobService {
	return &JobService{tx: tx}
}

// --- Проверка 3 (позитив): анонимная tx-сигнатура в поле структуры ---

type BadService struct {
	tx func(ctx context.Context, fn func(ctx context.Context) error) error // want `GID-175: используйте именованный тип model.InTransactionFunc`
}

// --- Проверка 3 (позитив): анонимная tx-сигнатура в параметре конструктора ---

func NewBadService(tx func(ctx context.Context, fn func(ctx context.Context) error) error) *BadService { // want `GID-175: используйте именованный тип model.InTransactionFunc`
	return &BadService{tx: tx}
}

// --- Проверка 4 (позитив): tx-метод на сервисе ---

func (s *JobService) Transaction(ctx context.Context, fn func(ctx context.Context) error) error { // want `GID-175: репозиторий/сервис не оборачивает транзакцию методом`
	return s.tx(ctx, fn)
}

// --- Граничный кейс: метод с похожей, но другой сигнатурой (callback без ctx) — не флагуем ---

func (s *JobService) NotTransaction(ctx context.Context, fn func() error) error {
	return nil
}
