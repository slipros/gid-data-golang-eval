// Eval of GID-272 non-applicability: a _test.go file is not judged — a test
// fixture may take as many arguments as the shape it builds needs (GID-250).
package model

func buildFixture(a, b, c, d, e int) *Model { return &Model{} }
