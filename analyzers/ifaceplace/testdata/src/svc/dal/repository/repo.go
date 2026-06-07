// Граничный класс GID-134: потребитель — слой /dal/repository.
// model-интерфейс здесь НЕ разрешён (исключение действует только для
// service/usecase).
package repository

import "svc/domain/model"

// Поле с model-интерфейсом в repository — нарушение.
type Repo struct {
	repo model.JobRepository // want `GID-134: interface JobRepository is declared in svc/domain/model\. Fix: define the interface next to its consumer \(exceptions: libraries and /domain/model for service/usecase\)`
}

// Параметр с model-интерфейсом в repository — нарушение.
func (r *Repo) Use(jr model.JobRepository) {} // want `GID-134: interface JobRepository is declared in svc/domain/model`
