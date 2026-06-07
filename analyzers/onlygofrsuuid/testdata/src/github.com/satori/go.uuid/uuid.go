// Stub запрещённой библиотеки для eval.
package uuid

type UUID [16]byte

func NewV4() UUID { return UUID{} }
