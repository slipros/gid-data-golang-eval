// Негатив: обычные импорты domain-слоя не задеваются.
package usecase

import "svc/domain/model"

type Upload struct{}

func (u *Upload) Snapshot() model.Snapshot {
	return model.Snapshot{}
}
