// The code under test.
package svc

func helper() int { return 42 }

// Value is the exported surface.
func Value() int { return helper() }
