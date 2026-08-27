// Eval for GID-271, negative class: the dependencies.go convention of lk-api
// and consent-api — a port file whose interfaces are consumed by several
// structs of the package (2–5 interfaces, 2–9 consumers). The owner allowed
// exactly this case ("можно выносить интерфейсы отдельно в файл, если они
// используются в нескольких сущностях") — the rule must stay silent here.
package repository

import "svc/domain/model"

// JobReader is used by Store and Archive — 2 consumers.
type JobReader interface {
	Job(id string) (model.Job, error)
}

// JobWriter is used by Store — overlaps with the other interfaces.
type JobWriter interface {
	Save(j model.Job) error
}

// JobLister is used by Archive and Remote — 2 consumers.
type JobLister interface {
	List() ([]model.Job, error)
}
