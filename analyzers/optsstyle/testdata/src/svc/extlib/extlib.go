// Package extlib stands in for an external dependency (another package): its
// Options type must never be flagged at the use site — it cannot be fixed there.
package extlib

type Options struct {
	Addr string
}
