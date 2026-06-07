// Eval для GID-104 (New<Entity>).
package service

type Snapshot struct{}

// Позитив: голый New.
func New() *Snapshot { // want `GID-104: конструктор именуется New<Entity> — голый New конфликтует с другими сущностями пакета`
	return &Snapshot{}
}

// Негатив: канонический конструктор.
func NewSnapshot() *Snapshot {
	return &Snapshot{}
}

// Неприменимость: метод New у типа — не package-level конструктор.
type Factory struct{}

func (f *Factory) New() *Snapshot { return &Snapshot{} }
