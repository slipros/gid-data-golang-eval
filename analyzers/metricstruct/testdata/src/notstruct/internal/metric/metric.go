// Prometheus is declared but is not a struct — violation.
package metric

// Prometheus — not a struct (a wrapper type); it even has Register.
type Prometheus int // want `GID-174: Prometheus must be a metrics aggregator struct\. Fix: make it a struct`

// Register exists, but the type is not a struct — still a violation.
func (p Prometheus) Register() error { return nil }
