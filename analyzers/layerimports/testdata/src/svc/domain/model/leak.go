// Positive (GID-227): model imports transport — forbidden,
// model is the pure vocabulary of the service.
package model

import "svc/server/middleware" // want `GID-227: package "svc/domain/model" must not import "svc/server/middleware"\. Fix: domain/model is the pure vocabulary of the service; layers do not flow into it`

// Legacy2 drags transport into the vocabulary — a violation.
type Legacy2 struct{}

func (l Legacy2) touch() {
	middleware.Noop()
}
