// Eval for GID-157 — order violations and interleaving.
package entitygroup

// Positive: a method above the type declaration.
func (s *Snapshot) Early() string { // want `GID-157: method "Early" must be placed below the "Snapshot" type declaration`
	return s.name
}

type Snapshot struct {
	name string
}

// Positive: the constructor ends up below a method of its entity.
func NewSnapshot(name string) *Snapshot { // want `GID-157: constructor "NewSnapshot" sits below the methods of "Snapshot"\. Fix: keep every constructor of an entity together under its type declaration, above the methods`
	return &Snapshot{name: name}
}

func (s *Snapshot) Name() string { return s.name }

type Job struct{}

func (j *Job) Run() error { return nil }

// Positive: a Snapshot method after the Job block — interleaving.
func (s *Snapshot) Render() string { // want `GID-157: entity "Snapshot" code is interleaved with other entities\. Fix: keep the entity block contiguous`
	return "<" + s.name + ">"
}
