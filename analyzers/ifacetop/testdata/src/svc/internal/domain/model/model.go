package model

type Order struct {
	ID int
}

type Identifiable interface {
	ID() int
}
