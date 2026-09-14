// Positive (GID-275): an alias of the module's own package that nothing in
// the file calls for.
package service

import (
	model "example.com/svc/internal/domain/model"                    // want `GID-275: import alias model is not needed — nothing else in the file is called model\. Fix: import it as "example\.com/svc/internal/domain/model" and refer to it as model`
	serviceconvert "example.com/svc/internal/domain/service/convert" // want `GID-275: import alias serviceconvert is not needed — nothing else in the file is called convert\. Fix: import it as "example\.com/svc/internal/domain/service/convert" and refer to it as convert`
	ev "example.com/svc/internal/event/v1"                           // want `GID-275: import alias ev is not needed — nothing else in the file is called eventv1\. Fix: import it as eventv1 "example\.com/svc/internal/event/v1", the real package name`
)

var _ = serviceconvert.X + model.X + ev.X
