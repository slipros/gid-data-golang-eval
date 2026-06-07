// Группа Kafka — отдельный файл.
package metric

// KafkaMetrics — группа Kafka.
type KafkaMetrics struct {
	Lag int
}

// Register группы Kafka.
func (m KafkaMetrics) Register() error { return nil }
