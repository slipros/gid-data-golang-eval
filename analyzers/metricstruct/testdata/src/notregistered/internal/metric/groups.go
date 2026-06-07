// Группа HTTP — отдельный файл.
package metric

// HTTPMetrics — группа HTTP.
type HTTPMetrics struct {
	Requests int
}

// Register группы HTTP.
func (m HTTPMetrics) Register() error { return nil }
