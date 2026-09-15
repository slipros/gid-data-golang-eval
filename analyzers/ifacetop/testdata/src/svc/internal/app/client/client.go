package client

type Wiring struct {
	c WiringCloser
}

type WiringCloser interface {
	Close() error
}
