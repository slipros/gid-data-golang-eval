// Not applicable: a leaf package without imports from its own module — the
// rule does not fire (no parent import).
package child

// Leaf — a type of the child package.
type Leaf struct {
	Name string
}
