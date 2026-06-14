// A local flag package — a namesake of stdlib, but with a different import path.
package flag

func String(name string) *string { return &name }

func Parse() {}
