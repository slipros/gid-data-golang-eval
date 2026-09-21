package model

type Order struct {
	ID     string
	Status string
	Title  string
}

type CreateOrder struct {
	Title  string
	Status string
}
