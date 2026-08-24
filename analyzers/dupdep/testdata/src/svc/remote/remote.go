// Package remote — an interface declared outside the constructor's package.
package remote

// Auditor — the audit dependency, owned by this package.
type Auditor interface {
	Audit(event string)
}
