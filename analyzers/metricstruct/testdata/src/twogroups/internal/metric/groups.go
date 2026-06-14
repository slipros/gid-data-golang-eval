// Check 7: two functional groups in one file — report on the second one.
package metric

// HTTPMetrics — the first group, ok.
type HTTPMetrics struct {
	Requests int
}

// Register of the HTTP group.
func (m HTTPMetrics) Register() error { return nil }

// GRPCMetrics — the second group in the same file — violation.
type GRPCMetrics struct { // want `GID-174: one functional metrics group per file\. Fix: split groups into separate files`
	Calls int
}

// Register of the GRPC group.
func (m GRPCMetrics) Register() error { return nil }
