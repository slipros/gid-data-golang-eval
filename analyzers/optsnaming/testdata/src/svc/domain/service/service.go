// Eval for GID-126: positive and boundary cases outside the app layer (/domain/service).
package service

import "context"

// --- Positive: a struct type named exactly Options outside the app layer ---

type Options struct { // want `GID-126: an options type must have an entity prefix\. Fix: use JobOptions, not bare Options`
	Retries int
}

// An entity-prefixed Options type — used by the defaults and parameters below.
type JobOptions struct {
	Retries int
}

// --- Positive: a package-level var of type <X>Options without the Default prefix ---

var Opts = JobOptions{Retries: 3} // want `GID-126: option defaults must be a Default<X>Options variable\. Fix: rename it`

// --- Positive: a package-level var declaration (explicit type) without Default ---

var defaults JobOptions // want `GID-126: option defaults must be a Default<X>Options variable\. Fix: rename it`

// --- Negative: defaults in a Default<X>Options variable ---

var DefaultJobOptions = JobOptions{Retries: 5}

// --- Boundary: a local variable opts — not matched ---

func use() int {
	opts := JobOptions{Retries: 1}
	return opts.Retries
}

// --- Boundary: a function with an opts parameter — not this rule's domain ---

func New(ctx context.Context, opts *JobOptions) int {
	_ = ctx
	return opts.Retries
}

// --- Boundary: a pointer var with the Default prefix — ok ---

var DefaultGRPCOptions *GRPCOptions

type GRPCOptions struct {
	Addr string
}
