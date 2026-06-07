// Eval GID-171: фильтр прямо в model-слое — ок.
package model

// --- Негативный класс: чистый код проходит ---

type JobsFilter struct {
	Status string
}

type FilterJobs struct {
	Limit int
}
