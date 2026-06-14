// Check 6: prometheus.go declares another exported struct group.
package metric

// Prometheus — the aggregator; it belongs here.
type Prometheus struct {
	HTTP HTTPMetrics
}

// HTTPMetrics in prometheus.go — violation: a group lives in its own file.
type HTTPMetrics struct { // want `GID-174: a metrics group must live in its own file; prometheus.go is wiring only\. Fix: move the group out`
	Requests int
}

// Register registers the HTTP group.
func (p Prometheus) Register() error { return p.HTTP.Register() }

// Register of the HTTP group.
func (m HTTPMetrics) Register() error { return nil }
