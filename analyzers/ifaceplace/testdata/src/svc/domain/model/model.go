// Package model — слой /domain/model. Интерфейсы отсюда разрешены
// потребителям service/usecase.
package model

// Job — обычная сущность model.
type Job struct {
	ID string
}

// JobRepository — интерфейс зависимости, объявленный в model-слое.
// Разрешён в service и usecase, запрещён в остальных слоях.
type JobRepository interface {
	Job(id string) (Job, error)
}
