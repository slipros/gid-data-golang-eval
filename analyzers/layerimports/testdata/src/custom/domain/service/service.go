// Positive (settings.rules): the custom rule SVC-1 bans legacy in domain/service.
package service

import "custom/legacy/store" // want `SVC-1: package "custom/domain/service" must not import "custom/legacy/store"\. Fix: the legacy package is being removed`

type Snapshot struct {
	st *store.Store
}
