// Eval of GID-246 negative: a struct without "adapter" in its name is never
// flagged, regardless of methods. This package is the composition root (app/api).
package api

// GRPCOptions is configuration data — no "adapter" in the name, not flagged.
type GRPCOptions struct {
	Host string
	Port int
}

// Addr is a convenience accessor — the rule does not look at methods at all.
func (o GRPCOptions) Addr() string { return o.Host }
