// Package config here is used by TestCustomScope with Settings{Leaf:[["config"]]}:
// under that custom scope the package IS in scope, so embedding is flagged.
package config

type Options struct {
	X int
}

type Root struct {
	Options // want `GID-152: embedding Options is forbidden`
}
