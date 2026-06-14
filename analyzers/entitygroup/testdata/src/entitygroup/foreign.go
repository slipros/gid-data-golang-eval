// Positive: an entity's method and constructor in a foreign file.
package entitygroup

func NewJob() *Job { // want `GID-157: "NewJob" belongs to entity "Job"\. Fix: keep the entity's code in the file where it is declared`
	return &Job{}
}

func (s *Snapshot) Foreign() string { // want `GID-157: "Foreign" belongs to entity "Snapshot"\. Fix: keep the entity's code in the file where it is declared`
	return s.name
}
