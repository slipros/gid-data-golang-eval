// Positive (GID-226): metric imports domain — forbidden,
// the Prometheus aggregator is standalone.
package metric

import "svc/domain/model" // want `GID-226: package "svc/metric" must not import "svc/domain/model"\. Fix: the metric package is a standalone Prometheus aggregator; service layers are not available to it`

type Prometheus struct{}

func (p *Prometheus) observe(in model.Snapshot) {}
