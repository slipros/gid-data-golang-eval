// Eval of GID-171: the model/filter subpackage is a full-fledged model layer, fine.
package filter

// --- Negative class: clean code passes ---

type JobsFilter struct {
	Status string
}

type Filter struct {
	Limit int
}
