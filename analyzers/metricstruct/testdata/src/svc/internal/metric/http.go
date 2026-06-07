// Функциональная группа HTTP-метрик — отдельный файл.
package metric

// HTTPMetrics — группа метрик HTTP-подсистемы.
type HTTPMetrics struct {
	Requests int
}

// Register регистрирует HTTP-метрики.
func (m HTTPMetrics) Register() error { return nil }
