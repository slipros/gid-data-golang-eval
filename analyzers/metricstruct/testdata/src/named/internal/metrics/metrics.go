// Путь оканчивается metrics — нарушение имени пакета (репорт на package clause).
package metric // want `GID-174: пакет метрик называется metric, не metrics`

// Prometheus тут есть и корректен, но путь metrics всё равно нарушение.
type Prometheus struct {
	HTTP int
}

// Register корректен.
func (p Prometheus) Register() error { return nil }
