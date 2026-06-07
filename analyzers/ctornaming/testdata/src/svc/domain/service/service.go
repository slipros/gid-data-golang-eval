// Eval для GID-104 (New<Entity>).
package service

type Snapshot struct{}

// Позитив: голый New.
func New() *Snapshot { // want `GID-104: a constructor must be named New<Entity>, not bare New\. Fix: rename it to New<Entity> \(bare New clashes with other entities in the package\)`
	return &Snapshot{}
}

// Негатив: канонический конструктор.
func NewSnapshot() *Snapshot {
	return &Snapshot{}
}

// Неприменимость: метод New у типа — не package-level конструктор.
type Factory struct{}

func (f *Factory) New() *Snapshot { return &Snapshot{} }
