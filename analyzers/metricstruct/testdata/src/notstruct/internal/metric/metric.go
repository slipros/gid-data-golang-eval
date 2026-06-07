// Prometheus объявлен, но не struct — нарушение.
package metric

// Prometheus — не struct (тип-обёртка), Register даже есть.
type Prometheus int // want `GID-174: Prometheus must be a metrics aggregator struct\. Fix: make it a struct`

// Register есть, но тип не struct — это всё равно нарушение.
func (p Prometheus) Register() error { return nil }
