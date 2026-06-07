// Eval GID-171: подпакет model/filter — полноправный model-слой, ок.
package filter

// --- Негативный класс: чистый код проходит ---

type JobsFilter struct {
	Status string
}

type Filter struct {
	Limit int
}
