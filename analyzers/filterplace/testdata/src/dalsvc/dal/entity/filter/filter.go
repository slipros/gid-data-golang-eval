// Eval GID-171: правильное место entity-фильтров — /dal/entity/filter.
package filter

// --- Негативный класс: чистый код проходит ---

// Фильтр в своём месте — диагностики нет.
type JobsFilter struct {
	Status string
}

type FilterStages struct {
	StageID string
}

// Голое имя Filter — тоже фильтр, но в своём месте — ок.
type Filter struct {
	Limit int
}
