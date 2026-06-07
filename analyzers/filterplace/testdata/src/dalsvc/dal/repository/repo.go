// Eval GID-171: фильтры в /dal/** вне /dal/entity/filter.
package repository

// --- Позитивный класс: нарушение ловится ---

// Суффикс *Filter — фильтр в repository, должен жить в /dal/entity/filter.
type JobsFilter struct { // want `GID-171: фильтр "JobsFilter" живёт в /dal/entity/filter`
	Status string
}

// Префикс Filter* — тоже фильтр.
type FilterStages struct { // want `GID-171: фильтр "FilterStages" живёт в /dal/entity/filter`
	StageID string
}

// --- Граничный класс ---

// FilterFunc — не struct (func-тип), правило не трогает.
type FilterFunc func(row string) bool

// Filterable — слово Filter с продолжением строчной буквой, не имя-фильтр.
type Filterable struct {
	Enabled bool
}

// Обычная сущность без слова Filter — не трогаем.
type Job struct {
	ID string
}
