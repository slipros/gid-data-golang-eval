// Eval for GID-104 (New<Entity>).
package service

type Snapshot struct{}

// Positive: bare New.
func New() *Snapshot { // want `GID-104: a constructor must be named New<Entity>, not bare New\. Fix: rename it to New<Entity> \(bare New clashes with other entities in the package\)`
	return &Snapshot{}
}

// Negative: the canonical constructor.
func NewSnapshot() *Snapshot {
	return &Snapshot{}
}

// Not applicable: a New method on a type is not a package-level constructor.
type Factory struct{}

func (f *Factory) New() *Snapshot { return &Snapshot{} }
