// Negative (GID-240): a common-prefixed alias satisfies the rule.
package service

import (
	commonservice "repo/internal/domain/service"
)

// Snapshot references the shared entity through the required commonservice alias.
type Snapshot struct {
	ref *commonservice.Snapshot
}
