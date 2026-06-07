// Позитив (settings.rules): своё правило SVC-1 банит legacy в domain/service.
package service

import "custom/legacy/store" // want `SVC-1: package "custom/domain/service" must not import "custom/legacy/store"\. Fix: the legacy package is being removed`

type Snapshot struct {
	st *store.Store
}
