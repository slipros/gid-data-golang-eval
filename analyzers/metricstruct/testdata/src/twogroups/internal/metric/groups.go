// Проверка 7: две функциональные группы в одном файле — репорт на второй.
package metric

// HTTPMetrics — первая группа, ок.
type HTTPMetrics struct {
	Requests int
}

// Register группы HTTP.
func (m HTTPMetrics) Register() error { return nil }

// GRPCMetrics — вторая группа в том же файле — нарушение.
type GRPCMetrics struct { // want `GID-174: одна функциональная группа метрик на файл`
	Calls int
}

// Register группы GRPC.
func (m GRPCMetrics) Register() error { return nil }
