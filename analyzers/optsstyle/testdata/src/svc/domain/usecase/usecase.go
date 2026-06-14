// Package usecase is in scope (path contains domain/usecase).
package usecase

type Options struct {
	N int
}

// Embedding a local Options type is a violation here too.
type UseCase struct {
	Options // want `GID-152: embedding Options is forbidden`
}
