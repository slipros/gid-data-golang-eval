// Eval GID-230 boundary: a "server" segment nested below another layer
// (client/connect/server/grpc/handler) must NOT be classified as the
// transport layer. pathseg.Contains would match "server" anywhere in the
// path, wrongly putting this package in scope; the anchored pathseg.HasLayer
// requires "server" to be the leading segment after the module root, so this
// package is out of scope and Purge below — which has no Handle method — is
// not flagged.
package handler

import "context"

type Purge struct{}

func (h *Purge) notHandle(ctx context.Context) error { return nil }
