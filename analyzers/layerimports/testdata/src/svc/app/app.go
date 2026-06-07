// Негатив (GID-225): app — composition root, ему доступны все слои.
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

// Wire собирает зависимости сервиса.
func Wire() *App {
	return &App{}
}
