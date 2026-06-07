// Eval для GID-193: неприменимость — пакет main не проверяется (bootstrap).
package main

// MainOptions в обычном пакете было бы нарушением, но пакет main исключён.
type MainOptions struct {
	Addr string
}

func main() {}
