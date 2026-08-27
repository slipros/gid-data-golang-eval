// Eval for GID-270 (model-place), part B: an exported struct without methods,
// constructors and the options suffix in /domain/service is a data model.
package service

import "mp/domain/model"

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

// --- Part C: no function of the layer hands out a struct declared here ---

// snapshot — an unexported data struct: parts A and B let it through as a
// package detail, part C judges the functions that hand it out.
type snapshot struct {
	rows []string
}

// tally — the same shape, handed out by a nested function literal.
type tally struct {
	total int
}

// runOptions — boundary: the options convention exempts the type in part C too.
type runOptions struct {
	retry int
}

// Snapshot — positive: a method returns a struct of this very package.
func (p *Processor) Snapshot() snapshot { // want `GID-270: method "Processor.Snapshot" returns "snapshot" — a struct declared in this package, and /domain/service holds no data types of its own\. Fix: declare the returned type in /domain/model \(or //nolint:gidmodelplace when explicitly intended\)`
	return snapshot{rows: p.queue}
}

// Snapshots — positive: a slice of the same struct is the same defect.
func (p *Processor) Snapshots() []snapshot { // want `GID-270: method "Processor.Snapshots" returns "snapshot"`
	return nil
}

// count — positive: a function literal nested in a body is judged like any
// other function.
func (p *Processor) count() int {
	sum := func() tally { // want `GID-270: function literal returns "tally"`
		return tally{total: len(p.queue)}
	}

	return sum().total
}

// WithQueue — negative: the entity of the layer is handed out by its own
// methods (Processor has behavior).
func (p *Processor) WithQueue(queue []string) *Processor {
	p.queue = queue

	return p
}

// Rebuild — negative: what a constructor assembles is the layer's entity.
func Rebuild(repo Repository) *Integration {
	return NewIntegration(repo)
}

// buildOptions — boundary: an options-suffixed name is settings, not cargo.
func buildOptions() runOptions {
	return runOptions{retry: 1}
}

// Timeout — negative: a foreign package's type is not this package's cargo.
func Timeout() model.User {
	return model.User{}
}
