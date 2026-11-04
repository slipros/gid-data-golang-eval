package app

import (
	"svc/dal/repository"
	"svc/domain/service"
)

func fixture() { _ = service.NewExport(repository.NewCampaign()) }
