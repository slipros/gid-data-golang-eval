package app

import "wiring/domain/service"

func testFixture() { _ = service.NewExport(service.NewCampaign()) }
