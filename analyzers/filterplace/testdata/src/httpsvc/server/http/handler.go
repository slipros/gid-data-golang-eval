// Eval GID-171: класс неприменимости — слой вне dal/domain.
package http

// --- Класс неприменимости: правило не действует вне dal/domain ---

// Фильтр в /server/http — правило не применяется, диагностики нет.
type JobsFilter struct {
	Status string
}

type FilterJobs struct {
	Limit int
}
