// Функциональная группа Kafka-метрик — отдельный файл.
package metric

// KafkaMetrics — группа метрик Kafka-подсистемы (Register на указателе).
type KafkaMetrics struct {
	Lag int
}

// Register регистрирует Kafka-метрики (pointer receiver).
func (m *KafkaMetrics) Register() error { return nil }
