// Eval для GID-157 — нарушения порядка и перемешивание.
package entitygroup

// Позитив: метод выше объявления типа и конструктора.
func (s *Snapshot) Early() string { // want `GID-157: метод "Early" размещается под объявлением типа "Snapshot"` `GID-157: метод "Early" размещается под конструктором NewSnapshot`
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
func (s *Snapshot) Render() string { // want `GID-157: код сущности "Snapshot" перемешан с кодом других сущностей — блок сущности непрерывен`
	return "<" + s.name + ">"
}
