// Eval of GID-171: filters in /dal/** outside /dal/entity/filter.
package repository

// --- Positive class: the violation is caught ---

// The *Filter suffix — a filter in repository, must live in /dal/entity/filter.
type JobsFilter struct { // want `GID-171: filter "JobsFilter" must live in /dal/entity/filter\. Fix: move it there`
	Status string
}

// The Filter* prefix — a filter too.
type FilterStages struct { // want `GID-171: filter "FilterStages" must live in /dal/entity/filter\. Fix: move it there`
	StageID string
}

// --- Boundary class ---

// FilterFunc — not a struct (a func type), the rule leaves it alone.
type FilterFunc func(row string) bool

// Filterable — the word Filter continued by a lowercase letter, not a filter name.
type Filterable struct {
	Enabled bool
}

// An ordinary entity without the word Filter — untouched.
type Job struct {
	ID string
}
