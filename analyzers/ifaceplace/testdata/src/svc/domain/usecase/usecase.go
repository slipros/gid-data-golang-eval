// Граничный класс GID-134: потребитель — слой /domain/usecase.
// model-интерфейс здесь разрешён (как и в service).
package usecase

import "svc/domain/model"

// Поле с model-интерфейсом в usecase — ОК.
type Usecase struct {
	repo model.JobRepository
}

// Параметр с model-интерфейсом в usecase — ОК.
func (u *Usecase) Use(jr model.JobRepository) {}
