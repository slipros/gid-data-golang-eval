// settings.prefix: "shared" replaces the default "common" prefix.
// Positive: the default-looking "commonservice" alias no longer satisfies
// the rule once the required prefix is "shared".
// Negative: "sharedservice" satisfies the custom prefix.
package usecase

import (
	commonservice "custom/internal/domain/service" // want `GID-240: import "custom/internal/domain/service" of shared internal entities must carry a shared-prefixed alias\. Fix: import it as sharedservice "custom/internal/domain/service"`
)

// Handler references the shared entity; the alias fails the custom prefix check.
type Handler struct {
	ref *commonservice.Snapshot
}
