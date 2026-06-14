// Package exclude — centralized rule exclusions via the linter settings in
// .golangci.yml. A list entry: "Method" (any type)
// or "Type.Method" (a specific type).
package exclude

// Match reports whether the recvType.method method is on the exclusion list.
func Match(list []string, recvType, method string) bool {
	for _, e := range list {
		if e == method || e == recvType+"."+method {
			return true
		}
	}
	return false
}
