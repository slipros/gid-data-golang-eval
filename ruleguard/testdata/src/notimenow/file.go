// Eval для GID-001 (no-time-now).
package notimenow

import "time"

type clock interface {
	Now() time.Time
}

// --- Позитивные кейсы: нарушение ловится ---

func bad() time.Time {
	return time.Now() // want `GID-001: используйте gdhelper\.StdTime\.Now\(\) вместо time\.Now\(\)`
}

// Граничный кейс: time.Now() как аргумент выражения.
func badNested() time.Duration {
	return time.Since(time.Now()) // want `GID-001: используйте gdhelper\.StdTime\.Now\(\) вместо time\.Now\(\)`
}

// --- Негативные кейсы: чистый код проходит ---

func good(c clock) time.Time {
	return c.Now()
}

// --- Неприменимость: другие функции пакета time не трогаем ---

func notApplicable() time.Time {
	return time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
}
