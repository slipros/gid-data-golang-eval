// Канонический wiring-файл: тип Prometheus + Register, регистрирующий группы.
package metric

// Prometheus агрегирует функциональные группы метрик.
type Prometheus struct {
	HTTP  HTTPMetrics
	Kafka *KafkaMetrics
	Total int // поле без метода Register — регистрировать не требуется
}

// Register регистрирует все группы.
func (p Prometheus) Register() error {
	if err := p.HTTP.Register(); err != nil {
		return err
	}
	return p.Kafka.Register()
}
