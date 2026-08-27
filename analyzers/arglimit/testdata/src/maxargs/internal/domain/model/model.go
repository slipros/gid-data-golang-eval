// Eval of GID-272 with a custom threshold: settings.max-args 1 allows a
// single argument, two is already a violation. The default (3) is overridden
// per-project when the team decides the boundary differently.
package model

import "context"

func one(a int) {}

func two(a, b int) {} // want `GID-272: function two takes 2 substantive arguments \(allowed: 1\)`

func threeWithCtx(ctx context.Context, a, b int) {} // want `GID-272: function threeWithCtx takes 2 substantive arguments \(allowed: 1\)`
