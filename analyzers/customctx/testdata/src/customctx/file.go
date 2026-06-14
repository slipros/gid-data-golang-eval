// Eval for GID-188 (a ban on custom context types).
package customctx

import (
	"context"
	"time"
)

// --- Class 1: positive (violations) ---

// Case 2: an interface embedding context.Context.
type MyContext interface { // want `GID-188: custom context type MyContext is forbidden\. Fix: pass context\.Context and store data via context\.WithValue \(helpers live in /domain/model, GID-165/166\)\.`
	context.Context
	Extra() string
}

// Case 1: a struct with the full set of context.Context methods.
type CtxStruct struct{} // want `GID-188: custom context type CtxStruct is forbidden`

func (CtxStruct) Deadline() (time.Time, bool) { return time.Time{}, false }
func (CtxStruct) Done() <-chan struct{}       { return nil }
func (CtxStruct) Err() error                  { return nil }
func (CtxStruct) Value(key any) any           { return nil }

// Case 3: the ctx parameter is a custom context type.
func useCustom(ctx MyContext) {} // want `GID-188: parameter ctx has type .*MyContext\. Fix: use context\.Context\.`

// --- Class 2: negative (clean code) ---

// A ctx parameter of the correct type.
func good(ctx context.Context) { _ = ctx }

// A struct with a single Done method — the method set does NOT cover context.Context.
type PartialCtx struct{}

func (PartialCtx) Done() <-chan struct{} { return nil }

// --- Class 3: edge cases ---

// interface { context.Context } — the embedding matches once (one diagnostic).
type OnlyEmbed interface { // want `GID-188: custom context type OnlyEmbed is forbidden`
	context.Context
}

// A type with Deadline/Done/Err/Value methods but different signatures — not context.Context.
type FakeCtx struct{}

func (FakeCtx) Deadline() string  { return "" }
func (FakeCtx) Done() bool        { return false }
func (FakeCtx) Err() string       { return "" }
func (FakeCtx) Value(key int) int { return 0 }

// A parameter named ctx but with the stdlib type — not a violation.
func boundaryGood(ctx context.Context, other FakeCtx) { _ = ctx; _ = other }
