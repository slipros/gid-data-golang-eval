package notests

// Merge — a third candidate, declared in a file that sorts after a.go: the
// per-package diagnostic is anchored to the first file by name, so it stays put
// when this file grows.
func Merge(left, right []string) []string {
	out := make([]string, 0, len(left)+len(right))
	for _, item := range left {
		out = append(out, item)
	}
	for _, item := range right {
		out = append(out, item)
	}

	return out
}
