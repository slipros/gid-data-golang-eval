// Eval for GID-004 (allptr).
package allptr

import gdhelper "gitlab.gid.team/gid-data/tech/golang/libs/helper.git"

type File struct {
	ID   string
	Name string
}

type Files []File

// --- Positive cases: the violation is caught ---

func bad(files []File) []string {
	var out []string
	for _, f := range files { // want `GID-004: ranging over a slice of structs copies each element\. Fix: range over gdhelper\.AllPtr\(items\)`
		out = append(out, f.Name)
	}
	return out
}

// Edge case: a named slice type.
func badNamed(files Files) []string {
	var out []string
	for _, f := range files { // want `GID-004: ranging over a slice of structs copies each element\. Fix: range over gdhelper\.AllPtr\(items\)`
		out = append(out, f.Name)
	}
	return out
}

// Edge case: the index is taken alongside the value — the value still copies.
func badIndexAndValue(files []File) []string {
	var out []string
	for i, f := range files { // want `GID-004: ranging over a slice of structs copies each element\. Fix: range over gdhelper\.AllPtr\(items\)`
		out = append(out, f.Name[:i])
	}
	return out
}

// --- Negative cases: clean code passes ---

func good(files []File) []string {
	var out []string
	for _, f := range gdhelper.AllPtr(files) {
		out = append(out, f.Name)
	}
	return out
}

// A slice of pointers — no copying, AllPtr is not needed.
func goodPtrSlice(files []*File) []string {
	var out []string
	for _, f := range files {
		out = append(out, f.Name)
	}
	return out
}

// --- Not applicable: no value variable is bound ---

// Index-only iteration copies nothing. AllPtr yields
// indices together with pointers. Numeric-index-only loops are left unchanged.
func notApplicableIndexOnly(files []File) int {
	n := 0
	for i := range files {
		n += i
	}
	return n
}

func notApplicableNoVars(files []File) int {
	n := 0
	for range files {
		n++
	}
	return n
}

// --- Not applicable: not slices of structs ---

func notApplicableStrings(names []string) int {
	n := 0
	for range names {
		n++
	}
	return n
}

func notApplicableMap(byID map[string]File) []string {
	var out []string
	for id := range byID {
		out = append(out, id)
	}
	return out
}

// Positive: index is used only to reach the struct, including mutation.
func indexedFields(files []File) {
	for i := range files { // want `GID-004: index-only loop accesses struct elements`
		files[i].Name = "changed"
	}
}

func indexedAddress(files []File) {
	for i := range files { // want `GID-004: index-only loop accesses struct elements`
		current := &files[i]
		current.Name = "changed"
	}
}

// Boundary: the numeric index and element replacement must be preserved.
func numericIndex(files []File) {
	for i := range files {
		files[i].Name = string(rune(i))
	}
}

func replaceElement(files []File) {
	for i := range files {
		files[i] = File{}
	}
}

func resizeSlice(files []File) {
	for i := range files {
		files[i].Name = "changed"
		files = files[:len(files)-1]
	}
}

func primitiveIndex(values []int) {
	for i := range values {
		values[i]++
	}
}
