// Non-applicability (GID-275): blank and dot imports bind no package name, and
// an import of another module keeps its own convention (gd prefix).
package usecase

import (
	. "example.com/svc/internal/domain/service"
	_ "example.com/svc/internal/domain/service/convert"

	gderrs "example.com/lib/errs"
)

var _ = X + gderrs.Code
