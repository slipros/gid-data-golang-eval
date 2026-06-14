// The functional group of Kafka metrics — a separate file.
package metric

// KafkaMetrics — the metrics group of the Kafka subsystem (Register on a pointer).
type KafkaMetrics struct {
	Lag int
}

// Register registers the Kafka metrics (pointer receiver).
func (m *KafkaMetrics) Register() error { return nil }
