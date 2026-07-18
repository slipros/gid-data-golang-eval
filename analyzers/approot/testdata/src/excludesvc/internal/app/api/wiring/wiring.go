// Eval of GID-246 settings.exclude: an exempted type name is not flagged, while
// other adapter-named structs in the same package still are.
package wiring

import "context"

// LegacyAdapter carries "adapter" but is listed in settings.exclude
// ("LegacyAdapter") — no diagnostic.
type LegacyAdapter struct {
	n int
}

func (b *LegacyAdapter) Do(ctx context.Context) error {
	_ = ctx
	b.n++
	return nil
}

// PaymentAdapter is not excluded — still flagged.
type PaymentAdapter struct { // want `GID-246: "PaymentAdapter" is an adapter struct`
	n int
}
