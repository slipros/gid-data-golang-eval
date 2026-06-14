// Check 5: Prometheus is declared outside prometheus.go.
package metric

// Prometheus in metric.go instead of prometheus.go — violation.
type Prometheus struct { // want `GID-174: the Prometheus aggregator must live in prometheus.go\. Fix: move it there`
	Total int
}

// Register exists (check 3 will not fire), no group fields (check 8 stays silent).
func (p Prometheus) Register() error { return nil }
