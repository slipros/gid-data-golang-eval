// Negative: the parent imports a child — the correct direction,
// there must be no diagnostic.
package parent

import "app/parent/child" // ok: the parent imports a child

// Root — the shared type of the parent package.
type Root struct {
	Leaf child.Leaf
}
