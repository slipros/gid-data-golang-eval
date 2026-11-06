package usecase

import gdhelper "gitlab.gid.team/gid-data/tech/golang/libs/helper.git"

type Five struct{ A, B, C, D, E int }

type Four struct{ A, B, C, D int }

var bigStruct = Five{A: 1, B: 2, C: 3, D: 4, E: 5} // want `GID-281:`
var smallStruct = Four{A: 1, B: 2, C: 3, D: 4}
var bigMap = map[int]int{1: 1, 2: 2, 3: 3, 4: 4} // want `GID-281:`
var smallMap = map[int]int{1: 1, 2: 2, 3: 3}
var goodMap = map[int]int{
	1: 1,
	2: 2,
	3: 3,
	4: 4,
}
var goodStruct = Five{
	A: 1,
	B: 2,
	C: 3,
	D: 4,
	E: 5,
}
var slice = []int{1, 2, 3, 4, 5, 6, 7, 8}

func eight(a, b, c, d, e, f, g, h int) {} // want `GID-281:`

func seven(a, b, c, d, e, f, g int) {}

func eightLines(
	a int,
	b int,
	c int,
	d int,
	e int,
	f int,
	g int,
	h int,
) {
}

func calls() {
	eight(1, 2, 3, 4, 5, 6, 7, 8) // want `GID-281:`
	seven(1, 2, 3, 4, 5, 6, 7)
	eightLines(
		1,
		2,
		3,
		4,
		5,
		6,
		7,
		8,
	)
}

func uniqueness(values []string) {
	seen := make(map[string]struct{}) // want `GID-286:`
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
	}
}

func exportedMap(values []string) map[string]struct{} {
	seen := make(map[string]struct{})
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
	}
	return seen
}

func otherMap(values []string) {
	counts := make(map[string]int)
	for _, value := range values {
		counts[value]++
	}
}

func existingHelperSet(values []string) {
	seen := gdhelper.NewSet[string]()
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
	}
}
