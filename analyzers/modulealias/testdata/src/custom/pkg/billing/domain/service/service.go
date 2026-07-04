// settings.prefix: "shared" — negative: "sharedservice" satisfies the custom prefix.
package service

import (
	sharedservice "custom/internal/domain/service"
)

// Snapshot references the shared entity through the required sharedservice alias.
type Snapshot struct {
	ref *sharedservice.Snapshot
}
