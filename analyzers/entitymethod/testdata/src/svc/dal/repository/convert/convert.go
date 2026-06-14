// Not applicable: the convert/ subpackage is out of scope (the scope is the layer root).
package convert

import "context"

type Snapshot struct{ ID string }

type Mapper struct{}

// A verb method without an entity and with a List prefix — but the package is out of scope, no diagnostic.
func (m *Mapper) ListSnapshots(ctx context.Context) ([]Snapshot, error) {
	return nil, nil
}
