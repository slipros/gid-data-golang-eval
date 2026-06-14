// Package model — the /domain/model layer. Interfaces from here are allowed
// for service/usecase consumers.
package model

// Job — an ordinary model entity.
type Job struct {
	ID string
}

// JobRepository — a dependency interface declared in the model layer.
// Allowed in service and usecase, forbidden in the other layers.
type JobRepository interface {
	Job(id string) (Job, error)
}
