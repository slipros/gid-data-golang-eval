package billing

type ChargeIn struct {
	Amount int
}

type Options struct {
	Addr string
}

type Core interface { // want `GID-276: interface Core is declared below type ChargeIn \(line 3\)`
	Do() error
}

type Client struct {
	core Core
}
