package service

import "context"

const defaultLimit = 10

var ErrLimit = limitError{}

type HelloRepository interface {
	Hello(ctx context.Context) error
}

type HelloMetrics interface {
	Inc()
}

type Hello struct {
	repo    HelloRepository
	metrics HelloMetrics
}

func NewHello(repo HelloRepository, metrics HelloMetrics) *Hello {
	return &Hello{repo: repo, metrics: metrics}
}

func (h *Hello) Hello(ctx context.Context) error {
	return h.repo.Hello(ctx)
}

type limitError struct{}

func (limitError) Error() string { return "limit" }
