// Register с pointer receiver и произвольной сигнатурой — ок (граничный).
package metric

import "context"

// Prometheus — агрегатор без полей-групп (метрики плоские).
type Prometheus struct {
	HTTP int
}

// Register с pointer receiver, параметрами и возвратом.
func (p *Prometheus) Register(ctx context.Context, namespace string) (int, error) {
	return 0, nil
}
