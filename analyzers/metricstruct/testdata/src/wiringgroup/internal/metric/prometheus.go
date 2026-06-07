// Проверка 6: в prometheus.go объявлена другая экспортируемая struct-группа.
package metric

// Prometheus — агрегатор, тут он уместен.
type Prometheus struct {
	HTTP HTTPMetrics
}

// HTTPMetrics в prometheus.go — нарушение: группа живёт в отдельном файле.
type HTTPMetrics struct { // want `GID-174: группа метрик живёт в отдельном файле — prometheus.go только wiring`
	Requests int
}

// Register регистрирует группу HTTP.
func (p Prometheus) Register() error { return p.HTTP.Register() }

// Register группы HTTP.
func (m HTTPMetrics) Register() error { return nil }
