// Проверка 5: Prometheus объявлен не в prometheus.go.
package metric

// Prometheus в файле metric.go вместо prometheus.go — нарушение.
type Prometheus struct { // want `GID-174: агрегатор Prometheus живёт в prometheus.go`
	Total int
}

// Register есть (проверка 3 не сработает), полей-групп нет (проверка 8 молчит).
func (p Prometheus) Register() error { return nil }
