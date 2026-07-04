// Eval for GID-236 settings (settings.suffixes / settings.exclude).
package service

import "context"

// JobStore — a dependency interface of another entity, matched only when
// settings.suffixes includes the custom "Store" suffix.
type JobStore interface {
	Job(ctx context.Context, id string) (string, error)
}

// JobRepository — a dependency interface of another entity, matched via the
// default "Repository" suffix (still active alongside the custom one).
type JobRepository interface {
	Job(ctx context.Context, id string) (string, error)
}

// --- Positive (settings.suffixes): the custom "Store" suffix flags a foreign entity too ---

type Upload struct {
	jobs JobStore // want `GID-236: service "Upload" uses repository "JobStore" of another entity`
}

// --- Non-applicability (settings.exclude: "LegacySnapshot"): the whole struct is skipped ---

type LegacySnapshot struct {
	jobs JobStore
}

// --- Non-applicability (settings.exclude: "Delivery.jobs"): only that field is skipped,
// other violations of the same struct are still caught ---

type Delivery struct {
	jobs   JobStore      // excluded pointwise via settings.exclude
	others JobRepository // want `GID-236: service "Delivery" uses repository "JobRepository" of another entity`
}
