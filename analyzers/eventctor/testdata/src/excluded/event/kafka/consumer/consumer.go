// Eval for GID-216: a constructor from settings.exclude is not flagged
// (inapplicability via settings).
package consumer

// LegacyConsumer — a consumer without a logger, but listed in settings.exclude.
type LegacyConsumer struct{}

func NewLegacyConsumer() *LegacyConsumer {
	return &LegacyConsumer{}
}
