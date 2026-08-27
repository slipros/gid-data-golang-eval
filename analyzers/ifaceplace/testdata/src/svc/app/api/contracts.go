// Eval for GID-271, boundary class: a port file with ZERO consumers — the
// interface is used nowhere in the package (only in signatures of other
// packages or as a public contract). The rule does not take this case on and
// stays silent.
package api

// ExternalGateway is a public contract of the package; no struct of this
// package has a field of this type.
type ExternalGateway interface {
	Call() error
}
