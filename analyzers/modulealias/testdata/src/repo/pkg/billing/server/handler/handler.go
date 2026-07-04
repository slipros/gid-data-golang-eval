// Positive/negative (GID-240): a dot-import still needs a common-prefixed
// alias; a blank import is a side-effect-only reference and is skipped.
package handler

import (
	. "repo/internal/domain/model" // want `GID-240: import "repo/internal/domain/model" of shared internal entities must carry a common-prefixed alias\. Fix: import it as commonmodel "repo/internal/domain/model"`

	_ "repo/internal/domain/service"
)

// Snapshot uses the dot-imported Shared type directly.
type Snapshot struct {
	m Shared
}
