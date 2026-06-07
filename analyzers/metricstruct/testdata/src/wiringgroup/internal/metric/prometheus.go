// Проверка 6: в prometheus.go объявлена другая экспортируемая struct-группа.
package metric

// Prometheus — агрегатор, тут он уместен.
type Prometheus struct {
	HTTP HTTPMetrics
}

// HTTPMetrics в prometheus.go — нарушение: группа живёт в отдельном файле.
type HTTPMetrics struct { // want `GID-174: a metrics group must live in its own file; prometheus.go is wiring only\. Fix: move the group out`
	Requests int
}

// Register регистрирует группу HTTP.
func (p Prometheus) Register() error { return p.HTTP.Register() }

// Register группы HTTP.
func (m HTTPMetrics) Register() error { return nil }
