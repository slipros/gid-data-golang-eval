package model

// Option — a same-module per-call option: a variadic parameter of this type
// does not make a call external.
type Option interface {
	name() string
}
