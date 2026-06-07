// Stub запрещённой библиотеки для eval.
package uuid

type UUID [16]byte

func New() UUID { return UUID{} }
