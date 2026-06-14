// Check 8: a group field with a Register method is not registered in
// Prometheus.Register.
package metric

// Prometheus aggregates the groups.
type Prometheus struct {
	HTTP  HTTPMetrics
	Kafka KafkaMetrics // want `GID-174: Prometheus.Register registers group Kafka\. Fix: call its Register`
	Total int          // boundary: a type without Register — no registration needed
}

// Register registers only HTTP; Kafka was forgotten.
func (p Prometheus) Register() error {
	return p.HTTP.Register()
}
