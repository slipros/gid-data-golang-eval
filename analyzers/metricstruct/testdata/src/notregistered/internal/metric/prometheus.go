// Проверка 8: поле-группа с методом Register не зарегистрирована в
// Prometheus.Register.
package metric

// Prometheus агрегирует группы.
type Prometheus struct {
	HTTP  HTTPMetrics
	Kafka KafkaMetrics // want `GID-174: Prometheus.Register регистрирует группу Kafka — вызовите её Register`
	Total int          // граничный: тип без Register — регистрировать не нужно
}

// Register регистрирует только HTTP, забыли Kafka.
func (p Prometheus) Register() error {
	return p.HTTP.Register()
}
