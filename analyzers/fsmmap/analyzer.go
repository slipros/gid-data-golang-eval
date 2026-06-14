// Package fsmmap implements rule GID-231 (fsm-map-unexported):
//
//   - GID-231 (gidfsmmap): in /domain/model an FSM transition map — a
//     package-level var whose type is map[E][]E, map[E]map[E]struct{} or
//     map[E]map[E]bool, where E is a string-based enum declared in the same
//     package — must be unexported. The map is an implementation detail of
//     the enum's CanTransitionTo method; consumers go through the method,
//     never through the map.
//
// An enum here is a named type with underlying string that has at least one
// const of that type in the same package (the same detection technique as in
// enumstring/enumplace). Key and value must refer to the same enum type:
// map[A][]B over two different enums is not a transition map. Maps over
// non-enum keys (int-based types, string types without consts, plain string)
// are never flagged. Requiring the map and the enum to live in the same file
// is intentionally out of scope (see fsmmap.feature).
//
// Source: model.md "Enum and State Machine (FSM)": the transition map is an
// unexported variable var <entity>StatusTransitions.
package fsmmap

import (
	"go/ast"
	"go/token"
	"go/types"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-231"

// Analyzer — rule GID-231: an FSM transition map in /domain/model must be unexported. Fix: var snapshotStatusTransitions = map[Status][]Status{...}.
var Analyzer = &analysis.Analyzer{
	Name: "gidfsmmap",
	Doc:  ruleID + ": an FSM transition map in /domain/model must be unexported. Fix: var snapshotStatusTransitions = map[Status][]Status{...}",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	// Scope: domain model layer only.
	if !pathseg.Contains(pass.Pkg.Path(), "domain", "model") {
		return nil, nil
	}

	enums := enumTypesWithConsts(pass)
	if len(enums) == 0 {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range vs.Names {
					checkVar(pass, name, enums)
				}
			}
		}
	}
	return nil, nil
}

// checkVar reports a package-level var whose type is a transition map over a
// local string enum if the var name is exported.
func checkVar(pass *analysis.Pass, name *ast.Ident, enums map[*types.Named]struct{}) {
	if name.Name == "_" || !name.IsExported() {
		return
	}
	v, ok := pass.TypesInfo.Defs[name].(*types.Var)
	if !ok {
		return
	}
	vType := v.Type()
	m, ok := vType.Underlying().(*types.Map)
	if !ok {
		return
	}
	enum, ok := enumKey(m.Key(), enums)
	if !ok {
		return
	}
	if !isTransitionValue(m.Elem(), enum) {
		return
	}
	pass.Reportf(name.Pos(),
		"%s: FSM transition map %s is exported. Fix: make it unexported: var %s = map[Status][]Status{...}",
		ruleID, name.Name, lowerFirst(name.Name))
}

// enumKey reports whether t is one of the package's string enums.
func enumKey(t types.Type, enums map[*types.Named]struct{}) (*types.Named, bool) {
	named, ok := t.(*types.Named)
	if !ok {
		return nil, false
	}
	if _, isEnum := enums[named]; !isEnum {
		return nil, false
	}
	return named, true
}

// isTransitionValue reports whether elem is a transition target collection
// over the same enum: []E, map[E]struct{} (empty struct) or map[E]bool.
func isTransitionValue(elem types.Type, enum *types.Named) bool {
	switch e := elem.Underlying().(type) {
	case *types.Slice:
		return types.Identical(e.Elem(), enum)
	case *types.Map:
		if !types.Identical(e.Key(), enum) {
			return false
		}
		elem := e.Elem()
		switch v := elem.Underlying().(type) {
		case *types.Struct:
			return v.NumFields() == 0
		case *types.Basic:
			return v.Kind() == types.Bool
		}
	}
	return false
}

// enumTypesWithConsts — named string types of the package having at least one
// const value (the detection technique from enumstring). An alias to string
// does not produce a *types.Named, so aliases are excluded by construction.
func enumTypesWithConsts(pass *analysis.Pass) map[*types.Named]struct{} {
	out := map[*types.Named]struct{}{}
	for _, obj := range pass.TypesInfo.Defs {
		c, ok := obj.(*types.Const)
		if !ok {
			continue
		}
		named, ok := c.Type().(*types.Named)
		if !ok {
			continue
		}
		obj := named.Obj()
		if obj.Pkg() != pass.Pkg {
			continue
		}
		basic, ok := named.Underlying().(*types.Basic)
		if !ok || basic.Kind() != types.String {
			continue
		}
		out[named] = struct{}{}
	}
	return out
}

// lowerFirst lowercases the first rune of s for the rename suggestion.
func lowerFirst(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToLower(r)) + s[size:]
}
