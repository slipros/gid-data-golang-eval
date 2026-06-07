// Группа GRPC — отдельный файл.
package metric

// GRPCMetrics — группа GRPC.
type GRPCMetrics struct {
	Calls int
}

// Register группы GRPC.
func (m GRPCMetrics) Register() error { return nil }
