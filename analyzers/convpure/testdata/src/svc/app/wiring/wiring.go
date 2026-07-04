// Non-applicability (GID-235): an ordinary package (not a convert package)
// may import a business layer freely — the rule only scopes convert packages.
package wiring

import "svc/domain/service"

func Wire(s *service.Snapshot) string {
	return s.ID
}
