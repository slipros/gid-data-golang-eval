// Package embedmutex implements rule GID-178 (gidembedmutex):
// a ban on embedding (an anonymous field) sync.Mutex / sync.RWMutex
// (as well as pointers to them) into structs.
//
// Embedding a mutex promotes its Lock/Unlock methods into the type's public
// API: external code can lock someone else's mutex. The mutex is kept as a
// named unexported field (mu sync.Mutex), remaining an implementation
// detail.
//
// Detection goes through go/types (pass.TypesInfo), not the selector text,
// to work reliably with aliased imports of the sync package. An anonymous
// struct field whose type, after stripping the pointer, is the named type
// Mutex or RWMutex from the standard "sync" package. A named field of any
// kind is allowed. Generated code is skipped.
package embedmutex

import (
	"go/ast"
	"go/types"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-178"

// Analyzer — rule GID-178: do not embed sync.Mutex/sync.RWMutex; use a named field (mu sync.Mutex). Fix: give the mutex a name.
var Analyzer = &analysis.Analyzer{
	Name:     "gidembedmutex",
	Doc:      ruleID + ": do not embed sync.Mutex/sync.RWMutex; use a named field (mu sync.Mutex). Fix: give the mutex a name",
	Requires: astwalk.Requires,
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	astwalk.NodesOf(pass, ast.IsGenerated, func(_ *ast.File, st *ast.StructType) {
		if st.Fields == nil {
			return
		}
		for _, field := range st.Fields.List {
			// An anonymous (embedded) field has no names.
			if len(field.Names) != 0 {
				continue
			}
			name, ok := embeddedMutexName(pass.TypesInfo.TypeOf(field.Type))
			if !ok {
				continue
			}
			pass.Reportf(field.Pos(),
				"%s: sync.%s is embedded in the struct. Fix: use a named mutex field (mu sync.Mutex), "+
					"otherwise Lock/Unlock leak into the type's API",
				ruleID, name)
		}
	})

	return nil, nil
}

// embeddedMutexName returns the mutex type name ("Mutex" or "RWMutex")
// if t (after stripping the pointer) is the named type Mutex/RWMutex from the
// "sync" package. Otherwise ok == false.
func embeddedMutexName(t types.Type) (string, bool) {
	if t == nil {
		return "", false
	}
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return "", false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	if pkg == nil || pkg.Path() != "sync" {
		return "", false
	}
	switch obj.Name() {
	case "Mutex", "RWMutex":
		return obj.Name(), true
	default:
		return "", false
	}
}
