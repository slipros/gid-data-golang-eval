// Eval для GID-157 — нарушения порядка и перемешивание.
package entitygroup

// Позитив: метод выше объявления типа и конструктора.
func (s *Snapshot) Early() string { // want `GID-157: method "Early" must be placed below the "Snapshot" type declaration` `GID-157: method "Early" must be placed below the NewSnapshot constructor`
	return s.name
}

type Snapshot struct {
	name string
}

func NewSnapshot(name string) *Snapshot {
	return &Snapshot{name: name}
}

func (s *Snapshot) Name() string { return s.name }

type Job struct{}

func (j *Job) Run() error { return nil }

// Позитив: метод Snapshot после блока Job — перемешивание.
func (s *Snapshot) Render() string { // want `GID-157: entity "Snapshot" code is interleaved with other entities\. Fix: keep the entity block contiguous`
	return "<" + s.name + ">"
}
