// Positive + negative:
//   - importing the parent "app/parent" from a child package is a GID-131 violation;
//   - importing the sibling "app/parent/other" — NOT a parent, no diagnostic.
package badchild

import (
	"app/parent"       // want `GID-131: a child package imports its parent app/parent\. Fix: invert the dependency, move shared code down and let the parent import children`
	"app/parent/other" // ok: a sibling, not a parent
)

// Up pulls types from the parent and sibling packages.
type Up struct {
	Root    parent.Root
	Sibling other.Sibling
}
