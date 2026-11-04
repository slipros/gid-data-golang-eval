package service

import "other/domain/service"

type Worker interface{ Work() }
type Campaign struct{}

func (*Campaign) Work()      {}
func NewCampaign() *Campaign { return &Campaign{} }

type Export struct{ worker Worker }

func NewExport(worker Worker) *Export { return &Export{worker: worker} }

type PublicExport struct{ Worker Worker }
type ExportOptions struct{ Worker Worker }

type CrossPackage struct {
	remote *remote.Worker // want `GID-148: service "CrossPackage" depends on service "Worker"`
}

func (*CrossPackage) build() {
	_ = remote.NewWorker() // want `GID-148: service "CrossPackage" constructs service "Worker" inside a method`
}

type Configured struct{ opts *ExportOptions }

func NewConfigured(opts *ExportOptions) *Configured { return &Configured{opts: opts} }
