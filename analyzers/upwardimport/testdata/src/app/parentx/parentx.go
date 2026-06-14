// Boundary: "app/parentx" is NOT a child of "app/parent" — the prefix is
// computed by path segments, not by string. The string "app/parent" is a
// string prefix of "app/parentx", but NOT a segment-wise one (no "app/parent/").
// There must be no diagnostic.
package parentx

import "app/parent" // ok: parentx is not a child of parent (segment-wise prefix)

// Holder uses a type from the non-parent package parent.
type Holder struct {
	Root parent.Root
}
