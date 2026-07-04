// Positive (GID-240): inside pkg/<module>, shared internal/** imports
// without a common-prefixed alias are forbidden.
package usecase

import (
	"repo/internal/domain/service" // want `GID-240: import "repo/internal/domain/service" of shared internal entities must carry a common-prefixed alias \(e\.g\. commonservice\)`

	svc "repo/internal/domain/model" // want `GID-240: import "repo/internal/domain/model" of shared internal entities must carry a common-prefixed alias \(e\.g\. commonservice\)`
)

// Handler references shared entities via non-conforming aliases.
type Handler struct {
	svcRef *service.Snapshot
	model  svc.Shared
}
