// Eval for GID-270 (model-place), part B: an exported struct without methods,
// constructors and the options suffix in /domain/service is a data model.
package service

// Repository — the service's dependency interface.
type Repository interface {
	Snapshot(id string) (string, error)
}

// TriggerBuild — positive: an exported data struct in /domain/service. The
// incident shape is consent-webhook-trigger's AltCraftTriggerV2Build
// (internal/domain/usecase/webhook_trigger_v2.go:405, 2026-08-27).
type TriggerBuild struct { // want `GID-270: data struct "TriggerBuild" is declared in /domain/service — it has no methods and is built by no constructor, so it is a data model\. Fix: move the type to /domain/model \(or //nolint:gidmodelplace when explicitly intended\)`
	WebhookID string
	Payload   []byte
	Attempt   int
}

// Permissions — positive, the second incident shape (lk-api
// pkg/integration/domain/service/permission.go:37, 2026-08-27).
type Permissions struct { // want `GID-270: data struct "Permissions" is declared in /domain/service`
	CanRead   bool
	CanWrite  bool
	CanGrant  bool
	ExpiresAt int64
	Scope     string
}

// Processor — boundary: a struct with behavior is the layer's entity; the
// method keeps it out of the rule.
type Processor struct {
	queue []string
}

// Run gives Processor behavior.
func (p *Processor) Run() {}

// Integration — boundary: a struct a constructor assembles is the layer's
// entity itself.
type Integration struct {
	repo Repository
}

// NewIntegration builds the service.
func NewIntegration(repo Repository) *Integration {
	return &Integration{repo: repo}
}

// Pair — boundary: the lowercase new prefix counts as a constructor too.
type Pair struct {
	A, B string
}

func newPair(a, b string) *Pair {
	return &Pair{A: a, B: b}
}

// --- Boundary: the options convention keeps its names out of the rule ---

// ServerOptions — settings of the server.
type ServerOptions struct {
	Timeout int
}

// SendOption — the singular form is exempt too.
type SendOption struct {
	Retry int
}

// JobConfig — settings under another convention name.
type JobConfig struct {
	Cron string
}

// QueryParams — request parameters.
type QueryParams struct {
	Limit int
}

// SyncSettings — settings under yet another convention name.
type SyncSettings struct {
	Batch int
}

// draft — non-applicability: an unexported struct is a package detail.
type draft struct {
	title string
}
