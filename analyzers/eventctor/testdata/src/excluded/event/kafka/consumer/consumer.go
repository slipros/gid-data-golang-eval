// Eval для GID-216: конструктор из settings.exclude не флагается
// (неприменимость через настройки).
package consumer

// LegacyConsumer — consumer без logger, но числится в settings.exclude.
type LegacyConsumer struct{}

func NewLegacyConsumer() *LegacyConsumer {
	return &LegacyConsumer{}
}
