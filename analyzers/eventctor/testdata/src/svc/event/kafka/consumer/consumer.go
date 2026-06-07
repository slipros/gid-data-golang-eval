// Eval для GID-216: consumer-scope (event + consumer).
package consumer

import (
	"github.com/sirupsen/logrus"
	"registry"
)

type Service interface{ Do() }

// --- Позитив: consumer-конструктор без logger-параметра ---

type OrderConsumer struct {
	svc Service
}

func NewOrderConsumer(svc Service) *OrderConsumer { // want `GID-216: consumer принимает \*logrus\.Logger и собирает Entry с полями broker/consumer \(см\. event\.md\)`
	return &OrderConsumer{svc: svc}
}

// --- Негатив: consumer с *logrus.Logger ---

type PaymentConsumer struct {
	log *logrus.Entry
}

func NewPaymentConsumer(log *logrus.Logger) *PaymentConsumer {
	return &PaymentConsumer{log: log.WithField("consumer", "payment")}
}

// --- Негатив: consumer с *logrus.Entry ---

type RefundConsumer struct {
	log *logrus.Entry
}

func NewRefundConsumer(log *logrus.Entry) *RefundConsumer {
	return &RefundConsumer{log: log}
}

// --- Граничный: schema-функция возвращает тип чужого пакета — не конструктор ---

func NewOrderCreatedSchema() *registry.Schema {
	return &registry.Schema{}
}

// --- Граничный: неэкспортируемый хелпер — не конструктор ---

type helper struct{}

func newHelper() *helper {
	return &helper{}
}
