// Non-applicability (GID-111): a nested "client" segment below another layer
// (server-side connect interceptor at connect/client/interceptor) is NOT the
// client layer — the layer is anchored to the first segment after the module
// root ("connect"). The package is out of scope, so its by-value input and
// by-pointer output are not flagged.
package interceptor

import "context"

type Req struct{ ID string }

type Resp struct{ ID string }

type Interceptor struct{}

func (i *Interceptor) Do(ctx context.Context, in Req) (*Resp, error) {
	return &Resp{ID: in.ID}, nil
}
