// Пакет вне metric-пути: правило не применяется даже при наличии
// Prometheus без Register (неприменимость).
package service

// Prometheus без Register — но путь не metric, диагностики нет.
type Prometheus struct {
	HTTP int
}
