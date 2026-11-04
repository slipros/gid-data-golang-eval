package app

import (
	"wiring/dal/repository"
	"wiring/domain/service"
	"wiring/domain/service/convert"
	"wiring/domain/usecase"
)

func wire() {
	worker := service.NewCampaign()
	_ = service.NewExport(&convert.Converter{})
	_ = convert.NewConverter(&convert.Converter{})
	_ = service.NewExport(worker) // want `GID-148: service "Export" receives domain dependency "Campaign". Fix:`
	var port service.Worker = worker
	alias := port
	_ = service.NewExport(alias)                  // want `GID-148:`
	_ = service.NewExport(service.Worker(worker)) // want `GID-148:`
	_ = service.NewExport(usecase.NewExport())    // want `GID-148:`
	_ = &service.PublicExport{Worker: worker}     // want `GID-148:`
	_ = &service.ExportOptions{Worker: worker}
	_ = service.NewConfigured(&service.ExportOptions{})
	_ = service.NewExport(&repository.Campaign{})
}
func unknown(worker service.Worker) { _ = service.NewExport(worker) }
