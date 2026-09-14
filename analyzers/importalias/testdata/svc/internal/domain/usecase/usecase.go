// Negative (GID-275): two aliased convert packages justify each other —
// dropping either alias alone would leave the other looking redundant.
package usecase

import (
	repoconvert "example.com/svc/internal/dal/repository/convert"
	serviceconvert "example.com/svc/internal/domain/service/convert"
	eventv1 "example.com/svc/internal/event/v1" // the real name goimports writes for a path that does not suggest it
)

var _ = repoconvert.X + serviceconvert.X + eventv1.X
