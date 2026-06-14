// Not applicable: the package does not use context at all — no diagnostics.
package nocontext

type Handler struct{}

func (Handler) Serve(req string) string { return req }

// A parameter named ctx, but it is just a string and there is no context in the
// package — GID-188 only cares about named context-like types;
// there is no named non-stdlib type here, no diagnostic.
func process(ctx string) string { return ctx }
