// Package clean — negative and non-applicability fixtures of GID-262:
// comments that explain the code, and text that only looks like a document
// reference.
package clean

import (
	"context"
	"time"
)

// Resolver reads cabinets in one batch call per page: a per-item resolve would
// cost N requests to the registry, so the page size bounds the traffic.
type Resolver struct {
	timeout time.Duration
}

// Cabinets deduplicates the ids and drops the empty ones before the call — the
// registry answers 400 on an empty id, and a repeated id costs a round trip.
func (r *Resolver) Cabinets(ctx context.Context, ids []string) ([]string, error) {
	// The tenant boundary is applied by the caller: a foreign id must come back
	// as "not found", never as somebody else's data.
	return ids, nil
}

// Parse reads the timestamp in RFC3339 — the format the registry answers in.
func Parse(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

// Encoding stays UTF-8: the registry rejects anything else, and the HTTP/2
// transport carries the header verbatim.
const Encoding = "UTF-8"

//nolint:giddocref // legitimate: the exception granted by ARD Р-11 — a narrow consumer interface
func directive() bool { return true }

//go:generate mockery --name Registry --case snake

// Verify checks the signature; SkipVerify disables the check in the sandbox.
func Verify(skipVerify bool) bool { return !skipVerify }

// Retries is capped at 3: the fourth attempt outlives the client timeout.
const Retries = 3
