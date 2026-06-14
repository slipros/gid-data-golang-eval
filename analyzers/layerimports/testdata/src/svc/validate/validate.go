// Positive (GID-224): the validator reaches into dal — forbidden;
// negative: model is allowed for the validator.
package validate

import (
	"svc/dal/entity" // want `GID-224: package "svc/validate" must not import "svc/dal/entity"\. Fix: validators work only with domain/model and request types`

	"svc/domain/model"
)

type Snapshot struct{}

// Negative: model is fine.
func (v *Snapshot) Validate(in model.Snapshot) error {
	return nil
}

// Positive above: entity is not available to the validator.
func (v *Snapshot) leak(in entity.Snapshot) {}
