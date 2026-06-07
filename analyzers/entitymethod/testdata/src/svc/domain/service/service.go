// Eval для GID-114 (service): корневой пакет /domain/service — в scope.
package service

import "context"

type Session struct{ ID string }

// S — однобуквенная «сущность»: проверка 3 не применяется (служебное имя).
type S struct{}

// --- Позитив ---

func (s *Session) ListSessions(ctx context.Context) ([]Session, error) { // want `GID-114: без префикса List — множественное число: Jobs вместо ListJobs`
	return nil, nil
}

func (s *Session) SessionByID(ctx context.Context, id string) (Session, error) { // want `GID-114: без суффикса ByID — Job\(ctx, id\) вместо JobByID`
	return Session{}, nil
}

// --- Негатив ---

func (s *Session) Session(ctx context.Context, id string) (Session, error) {
	return Session{}, nil
}

func (s *Session) Sessions(ctx context.Context) ([]Session, error) {
	return nil, nil
}

// --- Граничный: однобуквенный ресивер S — имя сущности не проверяется ---

// Имя метода без «S», но сущность служебная (len <= 2) — диагностики нет.
func (x *S) Touch(ctx context.Context) error {
	return nil
}

// Префикс List всё равно ловится — не зависит от длины имени сущности.
func (x *S) ListAll(ctx context.Context) error { // want `GID-114: без префикса List — множественное число: Jobs вместо ListJobs`
	return nil
}
