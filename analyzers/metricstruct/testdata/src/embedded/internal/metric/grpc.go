// The GRPC group — a separate file.
package metric

// GRPCMetrics — the GRPC group.
type GRPCMetrics struct {
	Calls int
}

// Register of the GRPC group.
func (m GRPCMetrics) Register() error { return nil }
