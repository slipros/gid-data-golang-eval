package service

import "context"

type (
	CartRepository interface {
		Cart(ctx context.Context) error
	}
	CartMetrics interface {
		Inc()
	}
)

type (
	Cart struct {
		repo    CartRepository
		metrics CartMetrics
	}
	CartNotifier interface { // want `GID-276: interface CartNotifier is declared below type Cart \(line 15\)`
		Notify()
	}
)
