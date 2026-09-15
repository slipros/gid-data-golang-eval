package repository

import "context"

type HelloConnection interface {
	Exec(ctx context.Context) error
}

type Hello struct {
	conn HelloConnection
}

func NewHello(conn HelloConnection) *Hello {
	return &Hello{conn: conn}
}

type HelloTx interface { // want `GID-276: interface HelloTx is declared below type Hello \(line 9\)`
	Commit() error
}
