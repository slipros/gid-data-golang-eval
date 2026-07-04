// Eval for GID-216: consumer scope (event + consumer).
package consumer

import (
	"github.com/sirupsen/logrus"
	"registry"
)

type Service interface{ Do() }

// --- Positive: a consumer constructor without a logger parameter ---

type OrderConsumer struct {
	svc Service
}

func NewOrderConsumer(svc Service) *OrderConsumer { // want `GID-216: a consumer constructor must take \*logrus\.Logger and build an Entry with broker/consumer fields \(see event\.md\)\. Fix: add a logger \*logrus\.Logger parameter and build the Entry with WithField in the constructor`
	return &OrderConsumer{svc: svc}
}

// --- Negative: a consumer with *logrus.Logger ---

type PaymentConsumer struct {
	log *logrus.Entry
}

func NewPaymentConsumer(log *logrus.Logger) *PaymentConsumer {
	return &PaymentConsumer{log: log.WithField("consumer", "payment")}
}

// --- Negative: a consumer with *logrus.Entry ---

type RefundConsumer struct {
	log *logrus.Entry
}

func NewRefundConsumer(log *logrus.Entry) *RefundConsumer {
	return &RefundConsumer{log: log}
}

// --- Boundary: a schema function returns a foreign package type — not a constructor ---

func NewOrderCreatedSchema() *registry.Schema {
	return &registry.Schema{}
}

// --- Boundary: an unexported helper — not a constructor ---

type helper struct{}

func newHelper() *helper {
	return &helper{}
}
