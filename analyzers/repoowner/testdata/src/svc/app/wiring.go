package app

import (
	"svc/dal/repository"
	"svc/domain/service"
)

type Getter struct{}

func (Getter) Campaign() *repository.Campaign { return repository.NewCampaign() }
func wire() {
	repo := repository.NewCampaign()
	_ = service.NewCampaign(repo)
	_ = service.NewCampaign(repository.NewCampaign())
	_ = service.NewExport(repo)                     // want `GID-280: repository "svc/dal/repository.Campaign" is supplied to services "svc/domain/service.Campaign" and "svc/domain/service.Export". Fix:`
	_ = service.NewExport(repository.NewCampaign()) // want `GID-280:`
	var narrow service.Reader = repo
	alias := narrow
	_ = service.NewExport(alias)                // want `GID-280:`
	_ = service.NewExport(service.Reader(repo)) // want `GID-280:`
	_ = service.NewExport(&repository.Alias{})  // want `GID-280:`
	_ = &service.Export{Repo: repo}             // want `GID-280:`
	_ = service.NewExport(Getter{}.Campaign())  // want `GID-280:`
	_, _ = service.NewHistory(repository.NewHistory())
	_ = service.CampaignOptions{Repo: repo}
}

func unknown(repo service.Reader) { _ = service.NewExport(repo) }
func reassigned() {
	var repo service.Reader = repository.NewCampaign()
	repo = repository.NewHistory()
	_ = service.NewExport(repo)
}

func conditional(repo service.Reader, flag bool) {
	if flag {
		repo = repository.NewCampaign()
	}
	_ = service.NewExport(repo)
}
func rangeAssignment() {
	var repo service.Reader = repository.NewCampaign()
	for _, repo = range []service.Reader{repository.NewHistory()} {
	}
	_ = service.NewExport(repo)
}
