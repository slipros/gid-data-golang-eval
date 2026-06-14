// Boundary case: an embedded field with a Register method also requires a call.
package metric

// Prometheus embeds the HTTPMetrics group (embedded).
type Prometheus struct {
	HTTPMetrics // want `GID-174: Prometheus.Register registers group HTTPMetrics\. Fix: call its Register`
	GRPCMetrics // embedded and registered below — ok
}

// Register registers only GRPC via the embedded name; HTTP was forgotten.
func (p Prometheus) Register() error {
	return p.GRPCMetrics.Register()
}
