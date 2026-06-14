// The wiring file for the package exercising check 7.
package metric

// Prometheus aggregates the groups.
type Prometheus struct {
	HTTP HTTPMetrics
	GRPC GRPCMetrics
}

// Register registers both groups.
func (p Prometheus) Register() error {
	if err := p.HTTP.Register(); err != nil {
		return err
	}
	return p.GRPC.Register()
}
