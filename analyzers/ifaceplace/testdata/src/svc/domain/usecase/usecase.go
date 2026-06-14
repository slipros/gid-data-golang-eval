// Boundary class of GID-134: the consumer is the /domain/usecase layer.
// A model interface is allowed here (as in service).
package usecase

import "svc/domain/model"

// A field with a model interface in usecase — OK.
type Usecase struct {
	repo model.JobRepository
}

// A parameter with a model interface in usecase — OK.
func (u *Usecase) Use(jr model.JobRepository) {}
