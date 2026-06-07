// Негатив (GID-224): domain/model — единственный слой, доступный
// транспорту; импорт model в handler разрешён.
package handler

import "svc/domain/model"

type Snapshot struct{}

func (h *Snapshot) Get() model.Snapshot {
	return model.Snapshot{}
}
