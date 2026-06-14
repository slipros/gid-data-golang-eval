// Eval for GID-157 — order violations and interleaving.
package entitygroup

// Positive: a method above the type declaration and the constructor.
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

// Positive: a Snapshot method after the Job block — interleaving.
func (s *Snapshot) Render() string { // want `GID-157: entity "Snapshot" code is interleaved with other entities\. Fix: keep the entity block contiguous`
	return "<" + s.name + ">"
}
