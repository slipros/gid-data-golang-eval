// The HTTP group — a separate file.
package metric

// HTTPMetrics — the HTTP group.
type HTTPMetrics struct {
	Requests int
}

// Register of the HTTP group.
func (m HTTPMetrics) Register() error { return nil }
