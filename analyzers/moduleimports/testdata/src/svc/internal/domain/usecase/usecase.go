// Non-applicability: a package of the CORE itself is not a module package —
// the rule only judges packages under pkg/<module>/<layer>/**.
package usecase

import (
	"svc/internal/dal/entity"
	"svc/internal/dal/repository"
)

// Integration — a core usecase reaching into the core dal, which is GID-132's
// business, not this rule's.
type Integration struct {
	repo repository.Integration
	row  entity.Integration
}
