// Позитив (GID-226): metric импортирует domain — запрещено,
// агрегатор Prometheus самостоятелен.
package metric

import "svc/domain/model" // want `GID-226: пакету "svc/metric" запрещён импорт "svc/domain/model" — пакет metric — самостоятельный агрегатор Prometheus, слои сервиса ему недоступны`

type Prometheus struct{}

func (p *Prometheus) observe(in model.Snapshot) {}
