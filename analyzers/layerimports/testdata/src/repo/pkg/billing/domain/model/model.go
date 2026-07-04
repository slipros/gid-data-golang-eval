// Negative: a pure model with no dependencies, scoped to the billing
// application module (pkg/billing).
package model

// Invoice is the billing module's own vocabulary type.
type Invoice struct {
	ID string
}
