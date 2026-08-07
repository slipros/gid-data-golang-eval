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

// Index-only iteration copies nothing, and AllPtr yields pointers instead of
// indices — it cannot replace this loop.
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
