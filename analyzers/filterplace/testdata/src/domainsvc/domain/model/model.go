// Eval of GID-171: a filter directly in the model layer — fine.
package model

// --- Negative class: clean code passes ---

type JobsFilter struct {
	Status string
}

type FilterJobs struct {
	Limit int
}
