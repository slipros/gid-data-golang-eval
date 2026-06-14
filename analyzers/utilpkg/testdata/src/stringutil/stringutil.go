// Negative case: stringutil contains "util" as a suffix, but the whole
// package name does not match a blacklist entry — no match.
package stringutil

func Noop() {}
