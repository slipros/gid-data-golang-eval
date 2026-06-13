// Eval for GID-152 (opts style).
package optsstyle

type HelloOptions struct {
	Retries int
}

var DefaultHelloOptions = HelloOptions{Retries: 3}

// --- Violation: opts by value in a parameter ---

func NewBad(opts HelloOptions) int { // want `GID-152: opts must be passed by pointer\. Fix: use \*HelloOptions`
	return opts.Retries
}

// --- Violation: embedding opts-struct promotes its fields into the public API ---

type EmbeddedHello struct {
	HelloOptions // want `GID-152: embedding HelloOptions is forbidden: it promotes option fields into the public API\. Fix: use an unexported named field .opts HelloOptions.`
}

// Embedding via pointer is also a violation.
type EmbeddedPtrHello struct {
	*HelloOptions // want `GID-152: embedding HelloOptions is forbidden: it promotes option fields into the public API\. Fix: use an unexported named field .opts HelloOptions.`
}

// Exported named Options field is also a violation.
type ExportedOptsHello struct {
	Opts HelloOptions // want `GID-152: Options field "Opts" must be unexported\. Fix: rename to .opts HelloOptions.`
}

// --- OK: unexported named field ---

func NewGood(opts *HelloOptions) *GoodHello {
	return &GoodHello{opts: opts}
}

type GoodHello struct {
	opts *HelloOptions // unexported named field — OK
}

// Unexported by-value named field is also OK.
type GoodHelloVal struct {
	opts HelloOptions
}

// --- Not applicable: a type without the Options suffix ---

type Config struct {
	Retries int
}

func WithConfig(cfg Config) int { return cfg.Retries }

type Holder struct {
	cfg Config
}

// Embedding a non-Options type is not covered by this rule.
type EmbeddedConfig struct {
	Config
}
