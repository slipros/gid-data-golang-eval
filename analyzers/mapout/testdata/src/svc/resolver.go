// Eval of GID-257: a map received as a PARAMETER is not filled in — the
// function returns its result. The fixtures mirror the 2026-08-06 incident
// (advertising-api internal/domain/service/ad_cabinet_resolver.go): a chunk
// resolver whose whole output went into a map handed down from the caller.
package svc

// Cabinet is the value the fixtures resolve.
type Cabinet struct {
	ID    string
	Title string
}

// Resolver owns the methods below.
type Resolver struct{}

// --- Positive: the incident shape — the result of the call goes into a map
// the caller passed in, and the signature promises nothing. ---

func (r *Resolver) resolveChunk(chunk []string, result map[string]Cabinet) {
	for _, id := range chunk {
		result[id] = Cabinet{ID: id, Title: "t"} // want `GID-257: a map received as a parameter is filled in`
	}
}

// --- Positive: a counter accumulated into a caller's map. ---

func countInto(items []string, counts map[string]int) {
	for _, item := range items {
		counts[item]++ // want `GID-257: a map received as a parameter is filled in`
	}
}

// --- Positive: compound assignment is a write just as much as a plain one. ---

func sumInto(items []string, totals map[string]int) {
	for _, item := range items {
		totals[item] += 2 // want `GID-257: a map received as a parameter is filled in`
	}
}

// --- Positive: delete and clear mutate a map without an assignment. ---

func dropFrom(ids []string, cache map[string]Cabinet) {
	for _, id := range ids {
		delete(cache, id) // want `GID-257: a map received as a parameter is filled in`
	}
}

func resetAll(cache map[string]Cabinet) {
	clear(cache) // want `GID-257: a map received as a parameter is filled in`
}

// --- Negative: the function returns its result — the shape the rule asks for. ---

func (r *Resolver) resolveChunkClean(chunk []string) map[string]Cabinet {
	result := make(map[string]Cabinet, len(chunk))
	for _, id := range chunk {
		result[id] = Cabinet{ID: id, Title: "t"}
	}

	return result
}

// --- Negative: a map parameter that is only READ — data going in, not a
// result coming out. ---

func titlesOf(ids []string, known map[string]Cabinet) []string {
	titles := make([]string, 0, len(ids))
	for _, id := range ids {
		if cabinet, ok := known[id]; ok {
			titles = append(titles, cabinet.Title)
		}
	}

	return titles
}

func countKnown(known map[string]Cabinet) int {
	total := 0
	for range known {
		total++
	}

	return total
}

// --- Negative (write-only discriminator): a cycle guard in a recursive walk.
// The set is READ before it is written, which is what makes it state the caller
// lends rather than a result — returning it would make the walk worse. ---

func walk(node *Node, visited map[*Node]bool) int {
	if visited[node] {
		return 0
	}
	visited[node] = true

	total := 1
	for _, child := range node.Children {
		total += walk(child, visited)
	}

	return total
}

// Node is the graph the recursive walk above traverses.
type Node struct {
	Children []*Node
}

// --- Negative (write-only discriminator): memoization across calls — the
// cache is consulted first, filled second. ---

func titleOf(id string, cache map[string]string) string {
	if title, ok := cache[id]; ok {
		return title
	}

	title := "computed:" + id
	cache[id] = title

	return title
}

// --- Boundary: a map held in a struct field is not a parameter. ---

// Cache owns its map.
type Cache struct {
	items map[string]Cabinet
}

func (c *Cache) put(id string, cabinet Cabinet) {
	c.items[id] = cabinet
}

// --- Boundary: a map RECEIVER — a method on a named map type mutates its own
// value, which is what the type exists for. ---

// Registry is a named map type with methods.
type Registry map[string]Cabinet

func (r Registry) Put(id string, cabinet Cabinet) {
	r[id] = cabinet
}

// --- Boundary: writing to a LOCAL map built inside the function. ---

func buildLocal(ids []string) map[string]Cabinet {
	local := make(map[string]Cabinet, len(ids))
	for _, id := range ids {
		local[id] = Cabinet{ID: id}
	}

	return local
}

// --- Non-applicability: no map parameter at all. ---

func plain(ids []string) int {
	return len(ids)
}
