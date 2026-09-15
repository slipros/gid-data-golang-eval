package billing

type Metrics interface {
	Inc()
}

type RefundIn struct {
	Amount int
}
