// Boundary (GID-235): svc/client/event/audit is a client-layer package whose
// path happens to contain the "event" segment in the middle
// (client/event/audit), not the banned event layer itself (which would be
// event/... right after the module root). A substring-style Contains check
// would false-positive on the nested "event" segment; the layer-anchored
// check must stay clean here.
package convert

import "svc/client/event/audit"

func FromEntry(e audit.Entry) string {
	return e.ID
}
