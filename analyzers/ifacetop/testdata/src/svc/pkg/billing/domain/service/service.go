package service

type Payment struct {
	repo PaymentRepository
}

type PaymentRepository interface { // want `GID-276: interface PaymentRepository is declared below type Payment \(line 3\)`
	Payment() error
}
