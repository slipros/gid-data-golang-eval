// Register with a pointer receiver and an arbitrary signature — ok (boundary case).
package metric

import "context"

// Prometheus — an aggregator without group fields (flat metrics).
type Prometheus struct {
	HTTP int
}

// Register with a pointer receiver, parameters, and return values.
func (p *Prometheus) Register(ctx context.Context, namespace string) (int, error) {
	return 0, nil
}
