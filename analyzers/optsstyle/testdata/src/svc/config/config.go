// Package config is OUT of scope under the default settings (it is neither a
// handler leaf nor domain/service|usecase). The same constructs that would be
// violations elsewhere produce NO diagnostics here — this verifies path scoping.
package config

type Options struct {
	X int
}

type Root struct {
	Options          // no diagnostic — out of scope
	Opts    Options  // no diagnostic — out of scope
}

func Build(opts Options) {} // no diagnostic — out of scope
