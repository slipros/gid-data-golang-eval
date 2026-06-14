// Eval of GID-171: the inapplicability class — a layer outside dal/domain.
package http

// --- Inapplicability class: the rule does not apply outside dal/domain ---

// A filter in /server/http — the rule does not apply, no diagnostic.
type JobsFilter struct {
	Status string
}

type FilterJobs struct {
	Limit int
}
