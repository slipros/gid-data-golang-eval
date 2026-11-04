package excluded

import (
	"svc/dal/repository"
	"svc/domain/service"
)

func wire() {
	repo := repository.NewCampaign()
	_ = service.NewCampaign(repo)
	_ = service.NewExport(repo)
}
