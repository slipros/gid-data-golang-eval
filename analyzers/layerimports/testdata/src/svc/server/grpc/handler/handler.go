// Неприменимость: транспортный слой — правило направлений к нему
// пока не применяется (handler работает с model через сервис).
package handler

import "svc/domain/model"

type Snapshot struct{}

func (h *Snapshot) Get() model.Snapshot {
	return model.Snapshot{}
}
