// Wiring-файл для пакета с проверкой 7.
package metric

// Prometheus агрегирует группы.
type Prometheus struct {
	HTTP HTTPMetrics
	GRPC GRPCMetrics
}

// Register регистрирует обе группы.
func (p Prometheus) Register() error {
	if err := p.HTTP.Register(); err != nil {
		return err
	}
	return p.GRPC.Register()
}
