package trivialonly

// Status — an enum whose whole exported surface is trivial.
type Status string

// Snapshot — a model with getters only.
type Snapshot struct {
	id     string
	status Status
}

// String — an enum String (GID-124): one conversion, no decision.
func (s Status) String() string {
	return string(s)
}

// ID — a getter.
func (s *Snapshot) ID() string {
	return s.id
}

// Status — a getter.
func (s *Snapshot) Status() Status {
	return s.status
}

// NewSnapshot — a one-line constructor.
func NewSnapshot(id string) *Snapshot {
	return &Snapshot{id: id}
}
