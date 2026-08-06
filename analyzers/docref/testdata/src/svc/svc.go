// Package svc — fixtures of GID-262: comments referencing the development
// documentation instead of the code.
package svc

import "context"

// AdCabinetResolver resolves ad cabinets through resource-registry (ARD Р-11): a single // want `GID-262: comment references development documentation`
// batch call per page.
type AdCabinetResolver struct {
	registry Registry
}

// Registry — the consumer-side view of the registry.
type Registry interface {
	Cabinets(ctx context.Context, ids []string) ([]string, error)
}

// Cabinets collects unique cabinet ids before the call (@ФТ-11). // want `GID-262: comment references development documentation`
func (r *AdCabinetResolver) Cabinets(ctx context.Context, ids []string) ([]string, error) {
	// BACKLOG B-48: the guard rejects a write outside a transaction. // want `GID-262: comment references development documentation`
	unique := dedup(ids)

	return r.registry.Cabinets(ctx, unique)
}

// dedup drops repeated ids.
func dedup(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}

	return out
}

// Reason returns the reason code of the last history record (ARD §12, задача 29). // want `GID-262: comment references development documentation`
func Reason() string {
	// один вызов на страницу — поштучный резолв дал бы N запросов (задача 29) // want `GID-262: comment references development documentation`
	return "ok"
}

/* Batch reads the page in one call (ARD Р-11/К-3, задача 30). */ // want `GID-262: comment references development documentation`
func Batch() int                                                  { return 1 }

// VERIFY (@ФТ-32): read permission missing → 403, no registry call. // want `GID-262: comment references development documentation`
func Verify() bool { return true }

// Decision taken in коммит 34640e6 — the step order inside the transaction. // want `GID-262: comment references development documentation`
func Decision() bool { return true }

// Threshold is agreed after data is collected (open question PRD §8). // want `GID-262: comment references development documentation`
func Threshold() int { return 0 }
