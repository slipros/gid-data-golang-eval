// The canonical wiring file: the Prometheus type + Register that registers the groups.
package metric

// Prometheus aggregates the functional metrics groups.
type Prometheus struct {
	HTTP  HTTPMetrics
	Kafka *KafkaMetrics
	Total int // a field without a Register method — no registration required
}

// Register registers all groups.
func (p Prometheus) Register() error {
	if err := p.HTTP.Register(); err != nil {
		return err
	}
	return p.Kafka.Register()
}
