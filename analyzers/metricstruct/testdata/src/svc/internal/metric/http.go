// The functional group of HTTP metrics — a separate file.
package metric

// HTTPMetrics — the metrics group of the HTTP subsystem.
type HTTPMetrics struct {
	Requests int
}

// Register registers the HTTP metrics.
func (m HTTPMetrics) Register() error { return nil }
