// Eval для GID-216: producer-scope (event + producer).
package producer

import "github.com/sirupsen/logrus"

type Service interface{ Do() }

// --- Позитив: producer-конструктор с *logrus.Logger ---

type OrderProducer struct {
	log *logrus.Logger
}

func NewOrderProducer(log *logrus.Logger) *OrderProducer { // want `GID-216: producer не принимает logger — ошибки пробрасываются вызывающему коду; осознанное исключение — //nolint:gideventctor`
	return &OrderProducer{log: log}
}

// --- Негатив: producer без logger ---

type PaymentProducer struct {
	svc Service
}

func NewPaymentProducer(svc Service) *PaymentProducer {
	return &PaymentProducer{svc: svc}
}
