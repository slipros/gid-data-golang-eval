// Package service is in scope (path contains domain/service).
package service

// Options is declared in this package — local, so GID-152 applies.
type Options struct {
	Retries int
}

// --- OK: unexported named field ---

func New(opts *Options) *Service { // pointer param — OK
	return &Service{opts: opts}
}

type Service struct {
	opts *Options // unexported named field — OK
}

type ServiceVal struct {
	opts Options // unexported by-value named field — OK
}

// --- Violations ---

func NewBad(opts Options) int { // want `GID-152: opts must be passed by pointer`
	return opts.Retries
}

type Embedded struct {
	Options // want `GID-152: embedding Options is forbidden`
}

type EmbeddedPtr struct {
	*Options // want `GID-152: embedding Options is forbidden`
}

type Exported struct {
	Opts Options // want `GID-152: Options field "Opts" must be unexported`
}
