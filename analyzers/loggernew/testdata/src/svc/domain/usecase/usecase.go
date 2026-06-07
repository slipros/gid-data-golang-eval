// Eval GID-214: граничный — вызов New() одноимённого чужого пакета.
package usecase

import (
	logrus "svc/domain/usecase/fakelog"
)

// Граничный кейс: logrus здесь — другой пакет (import path
// svc/domain/usecase/fakelog). Резолв идёт по import-пути, а не по имени,
// поэтому диагностики нет.
func New() {
	_ = logrus.New()
}
