// Позитив (GID-226): metric импортирует domain — запрещено,
// агрегатор Prometheus самостоятелен.
package metric

import "svc/domain/model" // want `GID-226: package "svc/metric" must not import "svc/domain/model"\. Fix: the metric package is a standalone Prometheus aggregator; service layers are not available to it`

type Prometheus struct{}

func (p *Prometheus) observe(in model.Snapshot) {}
