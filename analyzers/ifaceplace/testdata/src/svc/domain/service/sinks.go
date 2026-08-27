// Eval for GID-271: two consumers of the sink_ports.go port file — the
// boundary between the violation (exactly 1) and the convention (2+).
package service

// First is the first consumer of Sink.
type First struct {
	sink Sink
}

// Second is the second consumer of Sink.
type Second struct {
	sink Sink
}
