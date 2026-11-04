package other

import (
	"svc/dal/repository"
	"svc/domain/service"
)

func independentPackage() { _ = service.NewExport(repository.NewCampaign()) }
