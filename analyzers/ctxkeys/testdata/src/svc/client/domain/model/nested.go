// Boundary (GID-165): this package lives at svc/client/domain/model — the
// domain/model segments are nested under the client layer, not anchored
// right after the module root, so it is NOT the domain/model layer itself.
// GID-165 (context.WithValue only in /domain/model) must still apply here.
// A substring-style Contains(path,"domain","model") check would
// false-positive by treating this package as the model layer and route it
// into the GID-166/167 helper-shape checks instead of the GID-165 ban; the
// layer-anchored check (HasLayer) must keep flagging the call below.
package model

import "context"

type localKey string

func Store(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, localKey("v"), v) // want `GID-165: context\.WithValue outside /domain/model is forbidden\. Fix: keep context keys and helpers in /domain/model so business layers do not depend on middleware`
}
