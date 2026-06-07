// Eval для settings.exclude: перечисленные методы не репортятся.
package service

import "context"

type Job struct{}

// Исключён как "Job.Close" (Тип.Метод) — иначе ловился бы проверкой 3.
func (j *Job) Close() error {
	return nil
}

// Исключён как "Ping" (имя метода).
func (j *Job) Ping(ctx context.Context) error {
	return nil
}

// Не исключён — ловится проверкой 3 (нет имени сущности Job).
func (j *Job) Flush(ctx context.Context) error { // want `GID-114: method name "Flush" must contain the entity name "Job"`
	return nil
}
