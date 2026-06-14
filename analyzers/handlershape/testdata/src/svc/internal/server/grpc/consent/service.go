// Eval GID-230: gRPC service struct shape (handlers as exported *Handler fields).
package consent

import (
	"svc/genproto/consentpb"
	"svc/internal/server/grpc/consent/handler"
)

// Consent — gRPC service struct: embeds Unimplemented*Server, exposes
// handlers as fields.
type Consent struct {
	consentpb.UnimplementedConsentServiceServer

	// Canonical: exported field with the Handler suffix.
	DocumentsHandler *handler.Documents

	// Unexported handler field.
	exportHandler *handler.Export // want `GID-230: gRPC service "Consent" must expose handlers as exported fields with the Handler suffix, got "exportHandler"\. Fix: DocumentsHandler \*handler\.Documents`

	// Exported but without the Handler suffix.
	Purge *handler.Purge // want `GID-230: gRPC service "Consent" must expose handlers as exported fields with the Handler suffix, got "Purge"\. Fix: DocumentsHandler \*handler\.Documents`
}

func (c *Consent) use() { _ = c.exportHandler } //nolint:unused

// Config — no Unimplemented*Server embed: not a gRPC service struct,
// arbitrary fields are fine (boundary).
type Config struct {
	Addr string
}
