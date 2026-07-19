// Eval for GID-216 with a custom settings.loggerTypes = ["mylog.Logger"]:
// the allowlist is authoritative — a project-specific logger type is accepted,
// while types outside the list (here slog.Logger) do NOT satisfy the rule.
package consumer

import (
	"log/slog"

	"mylog"
)

// --- Negative: the custom logger type is in the allowlist ---

type OrderConsumer struct {
	log *mylog.Logger
}

func NewOrderConsumer(log *mylog.Logger) *OrderConsumer {
	return &OrderConsumer{log: log.With("consumer", "order")}
}

// --- Positive: no logger at all ---

type PaymentConsumer struct{}

func NewPaymentConsumer() *PaymentConsumer { // want `GID-216: a consumer constructor must take a logger and enrich it with broker/consumer fields \(see event\.md\)\. Fix: add a logger parameter \(e\.g\. \*slog\.Logger\) and attach the broker/consumer fields in the constructor`
	return &PaymentConsumer{}
}

// --- Positive: slog.Logger is NOT in this custom allowlist, so it does not
// count as a logger here (proves the list, not a hardcoded set, decides) ---

type RefundConsumer struct {
	log *slog.Logger
}

func NewRefundConsumer(log *slog.Logger) *RefundConsumer { // want `GID-216: a consumer constructor must take a logger and enrich it with broker/consumer fields \(see event\.md\)\. Fix: add a logger parameter \(e\.g\. \*slog\.Logger\) and attach the broker/consumer fields in the constructor`
	return &RefundConsumer{log: log}
}
