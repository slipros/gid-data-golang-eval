// Negative (GID-225): app is the composition root, all layers are available to it.
package app

import (
	"svc/client/billing"
	"svc/dal/repository"
	"svc/domain/service"
	"svc/event/producer"
	"svc/metric"
)

type App struct {
	repo     *repository.Snapshot
	svc      *service.Snapshot
	client   *billing.Client
	producer *producer.Snapshot
	metric   *metric.Prometheus
}

// Wire assembles the service dependencies.
func Wire() *App {
	return &App{}
}
