// Eval for GID-216: producer scope (event + producer).
package producer

import "github.com/sirupsen/logrus"

type Service interface{ Do() }

// --- Positive: a producer constructor with *logrus.Logger ---

type OrderProducer struct {
	log *logrus.Logger
}

func NewOrderProducer(log *logrus.Logger) *OrderProducer { // want `GID-216: a producer constructor must not take a logger; errors are propagated to the caller\. Fix: remove the logger \(intentional exception: //nolint:gideventctor\)`
	return &OrderProducer{log: log}
}

// --- Negative: a producer without a logger ---

type PaymentProducer struct {
	svc Service
}

func NewPaymentProducer(svc Service) *PaymentProducer {
	return &PaymentProducer{svc: svc}
}
