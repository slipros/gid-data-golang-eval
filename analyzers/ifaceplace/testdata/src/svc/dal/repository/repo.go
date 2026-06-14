// Boundary class of GID-134: the consumer is the /dal/repository layer.
// A model interface is NOT allowed here (the exception applies only to
// service/usecase).
package repository

import "svc/domain/model"

// A field with a model interface in repository — a violation.
type Repo struct {
	repo model.JobRepository // want `GID-134: interface JobRepository is declared in svc/domain/model\. Fix: define the interface next to its consumer \(exceptions: libraries and /domain/model for service/usecase\)`
}

// A parameter with a model interface in repository — a violation.
func (r *Repo) Use(jr model.JobRepository) {} // want `GID-134: interface JobRepository is declared in svc/domain/model`
