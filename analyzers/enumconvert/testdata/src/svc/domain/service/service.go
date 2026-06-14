// Edge: the same construct outside a convert package — not matched.
// The service package (the last path segment is not convert) → out of scope.
package service

type (
	EntityStatus string
	ModelStatus  string
)

var statusMap = map[EntityStatus]ModelStatus{"active": "active"}

// The same enum map indexing without comma-ok — but outside a convert package, not matched.
func mapStatus(s EntityStatus) ModelStatus {
	return statusMap[s]
}
