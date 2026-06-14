// Not applicable: outside service/usecase/repository private package
// functions are the norm.
package util

func helper(s string) string { return s }

func Public(s string) string { return helper(s) }
