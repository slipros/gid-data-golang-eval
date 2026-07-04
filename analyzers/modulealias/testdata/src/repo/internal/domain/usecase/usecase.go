// Boundary: this package lives under repo/internal/**, not under
// pkg/<module> — GID-240 is out of scope here, so a plain (unaliased) import
// of another internal/** entity is fine.
package usecase

import "repo/internal/domain/service"

// Handler consumes the shared service entity without an alias — allowed
// outside pkg/<module>.
type Handler struct {
	ref *service.Snapshot
}
