// Eval of GID-171: filters in /domain/** outside the model layer.
package service

// --- Positive class: the violation is caught ---

// The Filter* prefix in service — must live in /domain/model.
type FilterJobs struct { // want `GID-171: filter "FilterJobs" must live in /domain/model\. Fix: move it there`
	Status string
}

// The *Filter suffix — a filter too.
type JobsFilter struct { // want `GID-171: filter "JobsFilter" must live in /domain/model\. Fix: move it there`
	Limit int
}

// --- Boundary class ---

// FilterFunc — a func type, not a struct, the rule leaves it alone.
type FilterFunc func(j string) bool

// Filterable — not a filter name (Filter + a lowercase letter).
type Filterable struct {
	On bool
}
