// Eval GID-171: фильтры в /domain/** вне model-слоя.
package service

// --- Позитивный класс: нарушение ловится ---

// Префикс Filter* в service — должен жить в /domain/model.
type FilterJobs struct { // want `GID-171: filter "FilterJobs" must live in /domain/model\. Fix: move it there`
	Status string
}

// Суффикс *Filter — тоже фильтр.
type JobsFilter struct { // want `GID-171: filter "JobsFilter" must live in /domain/model\. Fix: move it there`
	Limit int
}

// --- Граничный класс ---

// FilterFunc — func-тип, не struct, правило не трогает.
type FilterFunc func(j string) bool

// Filterable — не имя-фильтр (Filter + строчная).
type Filterable struct {
	On bool
}
