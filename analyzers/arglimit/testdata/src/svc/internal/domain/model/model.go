// Eval of GID-272: a function or method in /domain/** takes at most 3
// substantive arguments — a function with more stopped being a conversion and
// became assembly. The fixtures mirror the incident (2026-08-27,
// consent-webhook-trigger): WebhooksTriggersV2FromConsentEventV2 with six
// arguments, four of them maps keyed by the same organizationID.
package model

import "context"

// --- Negative: the allowed shapes — 0..3 arguments. ---

func zero() {}

func one(a int) int { return a }

func two(a, b int) {}

func three(a, b int, c string) {}

func threeNames(a, b, c int) {}

// --- Negative: context.Context is technical, not substantive — 3 arguments
// plus ctx stay 3 and pass. ---

func withCtx(ctx context.Context, a, b, c int) {}

// --- Boundary: exactly 3 is the maximum allowed; exactly 4 is already a
// violation (reported below under Positive). ---

func maxAllowed(a, b, c int) {}

// --- Positive: 4 arguments — the violation. ---

func four(a, b int, c, d string) {} // want `GID-272: function four takes 4 substantive arguments \(allowed: 3\)`

// --- Positive: a method counts the same as a function; the receiver is not
// an argument. ---

func (m *Model) methodFour(a, b, c, d int) {} // want `GID-272: method Model.methodFour takes 4 substantive arguments \(allowed: 3\)`

// --- Positive: the variadic tail is one argument. ---

func variadic(a, b, c int, rest ...string) {} // want `GID-272: function variadic takes 4 substantive arguments \(allowed: 3\)`

// --- Positive: many arguments of one kind — the incident shape (four maps
// keyed by the same id are one thing and ask to be grouped). ---

func incident( // want `GID-272: function incident takes 4 substantive arguments \(allowed: 3\)`
	organizationID string,
	triggers map[string][]Trigger,
	disabled map[string]bool,
	events map[string][]Event,
) {}

// --- Non-applicability: a constructor is not judged — it takes as many
// dependencies as its entity has. ---

func NewModel(a, b, c, d, e, f int) *Model { return &Model{} }

func newModel(a, b, c, d int) *Model { return &Model{} }

// --- Non-applicability: an unnamed parameter is still one argument. ---

func unnamed(string, int, string, int) {} // want `GID-272: function unnamed takes 4 substantive arguments \(allowed: 3\)`

// Model is the entity the fixtures build.
type Model struct {
	ID string
}

// Trigger is an event trigger.
type Trigger struct {
	ID string
}

// Event is an event payload.
type Event struct {
	ID string
}
