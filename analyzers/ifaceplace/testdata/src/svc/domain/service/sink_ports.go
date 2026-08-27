// Eval for GID-271, boundary class: a port file with exactly TWO consumers —
// the lower edge of the allowed convention ("используются в нескольких
// сущностях"). Exactly 1 is a violation (ports.go in usecase), exactly 2
// stays silent.
package service

// Sink is consumed by the two structs of sinks.go — the file is allowed.
type Sink interface {
	Flush() error
}
