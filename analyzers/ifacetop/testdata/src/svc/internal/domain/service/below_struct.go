package service

import "context"

type Order struct {
	repo   OrderRepository
	wallet OrderWalletService
}

type OrderRepository interface { // want `GID-276: interface OrderRepository is declared below type Order \(line 5\); interfaces open the file, right after import, const and var\. Fix: move type OrderRepository interface above type Order`
	Order(ctx context.Context) error
}

type OrderWalletService interface { // want `GID-276: interface OrderWalletService is declared below type Order \(line 5\)`
	Charge(ctx context.Context) error
}
