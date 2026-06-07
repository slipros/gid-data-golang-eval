// Eval для GID-133 (privatefunc).
package service

import "context"

type Snapshot struct {
	name string
}

// Конструктор тоже задаёт принадлежность: normalize ниже используется
// сущностями Snapshot (через конструктор) и Job — общий хелпер, норма.
func NewSnapshot() *Snapshot {
	return &Snapshot{name: normalize("snapshot")}
}

type Job struct{}

// --- Позитив: используется только одной сущностью ---

func (s *Snapshot) Render(ctx context.Context) string {
	return decorate(s.name) // единственный потребитель — Snapshot
}

func decorate(s string) string { // want `GID-133: приватная функция "decorate" используется только сущностью "Snapshot" — оформите её методом`
	return ">" + s
}

// --- Позитив: не используется никем ---

func orphan() string { // want `GID-133: приватная функция "orphan" принадлежит пакету — оформите её методом структуры`
	return ""
}

// --- Негатив: общий хелпер двух сущностей ---

func (j *Job) Name() string { return normalize("job") }

func normalize(s string) string {
	return s
}

// --- Неприменимость: экспортируемые функции и init не проверяются ---

func Shared(s string) string { return s }

func init() {}
