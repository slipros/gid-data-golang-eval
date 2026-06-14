// Eval of GID-171: the correct place for entity filters is /dal/entity/filter.
package filter

// --- Negative class: clean code passes ---

// A filter in its proper place — no diagnostic.
type JobsFilter struct {
	Status string
}

type FilterStages struct {
	StageID string
}

// The bare name Filter is a filter too, but in its proper place — fine.
type Filter struct {
	Limit int
}
