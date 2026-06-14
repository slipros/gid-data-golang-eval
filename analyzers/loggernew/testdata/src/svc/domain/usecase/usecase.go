// Eval GID-214: boundary — a call to New() of a foreign package with the same name.
package usecase

import (
	logrus "svc/domain/usecase/fakelog"
)

// Boundary case: logrus here is a different package (import path
// svc/domain/usecase/fakelog). Resolution is by import path, not by name,
// so there is no diagnostic.
func New() {
	_ = logrus.New()
}
