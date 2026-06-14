// Negative: ordinary domain-layer imports are not touched.
package usecase

import "svc/domain/model"

type Upload struct{}

func (u *Upload) Snapshot() model.Snapshot {
	return model.Snapshot{}
}
