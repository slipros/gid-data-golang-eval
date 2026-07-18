// Non-applicability (GID-229): a nested "client" segment below another layer
// (server-side connect interceptor at connect/client/interceptor) is NOT the
// client layer — the layer is anchored to the first segment after the module
// root ("connect"), so importing domain/model is not a violation here.
package interceptor

import "svc/domain/model"

func Wrap(in model.Snapshot) model.Snapshot { return in }
