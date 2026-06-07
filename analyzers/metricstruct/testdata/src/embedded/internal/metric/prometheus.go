// Граничный: embedded-поле с методом Register тоже требует вызова.
package metric

// Prometheus встраивает группу HTTPMetrics (embedded).
type Prometheus struct {
	HTTPMetrics // want `GID-174: Prometheus.Register registers group HTTPMetrics\. Fix: call its Register`
	GRPCMetrics // embedded и зарегистрирована ниже — ок
}

// Register регистрирует только GRPC через embedded-имя, забыли HTTP.
func (p Prometheus) Register() error {
	return p.GRPCMetrics.Register()
}
