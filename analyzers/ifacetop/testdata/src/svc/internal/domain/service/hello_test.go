package service

import "context"

type fakeHelloRepository struct{}

func (fakeHelloRepository) Hello(context.Context) error { return nil }

type helloHarness interface {
	Run()
}
