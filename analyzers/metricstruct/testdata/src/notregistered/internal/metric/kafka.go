// The Kafka group — a separate file.
package metric

// KafkaMetrics — the Kafka group.
type KafkaMetrics struct {
	Lag int
}

// Register of the Kafka group.
func (m KafkaMetrics) Register() error { return nil }
