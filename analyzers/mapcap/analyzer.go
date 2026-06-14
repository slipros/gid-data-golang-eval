// Package mapcap implements rule GID-183 (Uber perf: map capacity hints):
// if a map is created via make(map[K]V) WITHOUT a capacity argument and then,
// in the same function, is filled in a range loop over a collection of known
// length (slice, map, string), a capacity hint should be given:
// make(map[K]V, len(src)). The standard prealloc does not cover this case — it
// only handles slices.
//
// Pattern within a single function:
//  1. m := make(map[K]V)        // or var m = make(map[K]V) — without capacity;
//  2. for ... := range src {    // src is a slice/map/string
//     m[...] = ...          // unconditional index assignment to m
//     }
//
// The heuristic is conservative — we match only clearly safe cases:
//   - between make and the loop, m MUST NOT be used in any way (neither filled
//     outside the loop nor passed into a call): any mention of m cancels the
//     diagnostic, since by the time of the loop its length is already unknown
//     to the analyzer;
//   - range over a channel is NOT matched — a channel has no len, its size is
//     unknown in advance;
//   - an m[...] = ... assignment inside an if (conditional fill) in the loop body
//     is NOT matched — the real number of inserts is less than len(src), and the
//     hint may hurt.
//
// make with a capacity already specified (make(map[K]V, n)) is correct, not matched.
// Generated code (ast.IsGenerated) is skipped.
// LoadMode is TypesInfo (types are needed to tell a slice/map/string from a channel).
package mapcap

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-183"

// Analyzer — rule GID-183: make(map) without capacity when filled from range; give the len(src) hint. Fix: make(map[K]V, len(src)).
var Analyzer = &analysis.Analyzer{
	Name: "gidmapcap",
	Doc:  ruleID + ": make(map) without capacity when filled from range; give the len(src) hint. Fix: make(map[K]V, len(src))",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				return true
			}
			checkBlock(pass, fn.Body.List)
			return true
		})
	}
	return nil, nil
}

// checkBlock analyzes the statement sequence of a single block: it finds a
// make(map) without capacity and looks below, in the same block, for a range
// loop that fills the map. It recurses into nested blocks (bodies of if/for/...)
// so the pattern inside a nested block is caught too.
func checkBlock(pass *analysis.Pass, stmts []ast.Stmt) {
	for i, stmt := range stmts {
		// Descent into nested blocks — the pattern is local to its statement block.
		inspectNestedBlocks(pass, stmt)

		name, makeCall := mapMakeWithoutCap(pass, stmt)
		if name == "" {
			continue
		}
		obj := pass.TypesInfo.ObjectOf(declIdent(stmt, name))
		if obj == nil {
			continue
		}
		analyzeAfterMake(pass, stmts[i+1:], obj, makeCall)
	}
}

// analyzeAfterMake scans the statements after make. It emits (via Report) a
// diagnostic if the nearest use of m is a range loop that unconditionally fills
// m over a slice/map/string. Any other use of m before such a loop cancels the
// diagnostic.
func analyzeAfterMake(pass *analysis.Pass, rest []ast.Stmt, obj types.Object, makeCall *ast.CallExpr) {
	for _, stmt := range rest {
		rng, isRange := stmt.(*ast.RangeStmt)
		if isRange && fillsMapInRange(pass, rng, obj) {
			if !rangeOverKnownLen(pass, rng.X) {
				return // range over a channel/unknown — size is unknown.
			}
			if usesObjOutsideAssign(pass, rng.Body, obj) {
				return // conditional fill or other use of m in the body — not matched.
			}
			pass.Reportf(makeCall.Pos(),
				"%s: make without capacity while filling from range. Fix: make(map[K]V, len(src))",
				ruleID)
			return
		}
		// Any use of m before a suitable loop cancels the diagnostic.
		if usesObj(pass, stmt, obj) {
			return
		}
	}
}

// mapMakeWithoutCap determines that stmt is a map declaration via make without
// a capacity argument. It returns the variable name and the make call itself.
// Supports  m := make(map[K]V)  and  var m = make(map[K]V).
func mapMakeWithoutCap(pass *analysis.Pass, stmt ast.Stmt) (string, *ast.CallExpr) {
	var lhs ast.Expr
	var rhs ast.Expr

	switch s := stmt.(type) {
	case *ast.AssignStmt:
		if len(s.Lhs) != 1 || len(s.Rhs) != 1 {
			return "", nil
		}
		lhs, rhs = s.Lhs[0], s.Rhs[0]
	case *ast.DeclStmt:
		gen, ok := s.Decl.(*ast.GenDecl)
		if !ok || gen.Tok.String() != "var" || len(gen.Specs) != 1 {
			return "", nil
		}
		vs, ok := gen.Specs[0].(*ast.ValueSpec)
		if !ok || len(vs.Names) != 1 || len(vs.Values) != 1 {
			return "", nil
		}
		lhs, rhs = vs.Names[0], vs.Values[0]
	default:
		return "", nil
	}

	ident, ok := lhs.(*ast.Ident)
	if !ok {
		return "", nil
	}
	call, ok := rhs.(*ast.CallExpr)
	if !ok || !isMakeBuiltin(pass, call) {
		return "", nil
	}
	if len(call.Args) == 0 {
		return "", nil
	}
	if _, ok := call.Args[0].(*ast.MapType); !ok {
		return "", nil // make([]T, ...) / make(chan T) — not a map.
	}
	if len(call.Args) >= 2 {
		return "", nil // capacity already specified — correct.
	}
	return ident.Name, call
}

// declIdent returns the *ast.Ident of the declared variable from stmt by name.
func declIdent(stmt ast.Stmt, name string) *ast.Ident {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		if id, ok := s.Lhs[0].(*ast.Ident); ok && id.Name == name {
			return id
		}
	case *ast.DeclStmt:
		if gen, ok := s.Decl.(*ast.GenDecl); ok {
			if vs, ok := gen.Specs[0].(*ast.ValueSpec); ok {
				return vs.Names[0]
			}
		}
	}
	return nil
}

// fillsMapInRange reports that the range loop body has an m[...] = ... assignment
// where m is our object (at the top level of the loop body, unconditionally).
func fillsMapInRange(pass *analysis.Pass, rng *ast.RangeStmt, obj types.Object) bool {
	for _, stmt := range rng.Body.List {
		if isIndexAssignTo(pass, stmt, obj) {
			return true
		}
	}
	return false
}

// isIndexAssignTo: stmt is an assignment of the form m[key] = value, where m is obj.
func isIndexAssignTo(pass *analysis.Pass, stmt ast.Stmt, obj types.Object) bool {
	assign, ok := stmt.(*ast.AssignStmt)
	if !ok {
		return false
	}
	for _, lhs := range assign.Lhs {
		idx, ok := lhs.(*ast.IndexExpr)
		if !ok {
			continue
		}
		if id, ok := idx.X.(*ast.Ident); ok && pass.TypesInfo.ObjectOf(id) == obj {
			return true
		}
	}
	return false
}

// rangeOverKnownLen: the range source has a known length (slice, array, map,
// string). A channel has no len — false.
func rangeOverKnownLen(pass *analysis.Pass, x ast.Expr) bool {
	t := pass.TypesInfo.TypeOf(x)
	if t == nil {
		return false
	}
	switch u := t.Underlying().(type) {
	case *types.Slice, *types.Array, *types.Map:
		return true
	case *types.Basic:
		return u.Info()&types.IsString != 0
	case *types.Pointer:
		// *[N]T — a pointer to an array, range is allowed and the length is known.
		elem := u.Elem()
		_, isArr := elem.Underlying().(*types.Array)
		return isArr
	default:
		return false
	}
}

// usesObjOutsideAssign: in the loop body obj is used anywhere other than
// unconditional m[...] = ... assignments at the top level of the body. Any such
// use (conditional fill inside an if, reading m, passing m into a call) makes the
// real size unknown — we cancel the diagnostic.
func usesObjOutsideAssign(pass *analysis.Pass, body *ast.BlockStmt, obj types.Object) bool {
	for _, stmt := range body.List {
		if isIndexAssignTo(pass, stmt, obj) {
			continue // unconditional fill at the top level — expected.
		}
		if usesObj(pass, stmt, obj) {
			return true
		}
	}
	return false
}

// usesObj: a reference to obj occurs in an arbitrary node.
func usesObj(pass *analysis.Pass, node ast.Node, obj types.Object) bool {
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if found {
			return false
		}
		if id, ok := n.(*ast.Ident); ok && pass.TypesInfo.ObjectOf(id) == obj {
			found = true
			return false
		}
		return true
	})
	return found
}

// inspectNestedBlocks recursively runs checkBlock on the bodies of nested
// compound statements so the pattern inside them is analyzed too.
func inspectNestedBlocks(pass *analysis.Pass, stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		checkBlock(pass, s.List)
	case *ast.IfStmt:
		if s.Body != nil {
			checkBlock(pass, s.Body.List)
		}
		if s.Else != nil {
			inspectNestedBlocks(pass, s.Else)
		}
	case *ast.ForStmt:
		if s.Body != nil {
			checkBlock(pass, s.Body.List)
		}
	case *ast.RangeStmt:
		if s.Body != nil {
			checkBlock(pass, s.Body.List)
		}
	case *ast.SwitchStmt:
		if s.Body != nil {
			checkBlock(pass, s.Body.List)
		}
	case *ast.TypeSwitchStmt:
		if s.Body != nil {
			checkBlock(pass, s.Body.List)
		}
	case *ast.CaseClause:
		checkBlock(pass, s.Body)
	case *ast.SelectStmt:
		if s.Body != nil {
			checkBlock(pass, s.Body.List)
		}
	case *ast.CommClause:
		checkBlock(pass, s.Body)
	case *ast.LabeledStmt:
		inspectNestedBlocks(pass, s.Stmt)
	}
}

// isMakeBuiltin: the call is the built-in make, not a local make function.
func isMakeBuiltin(pass *analysis.Pass, call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "make" {
		return false
	}
	builtin, ok := pass.TypesInfo.Uses[ident].(*types.Builtin)
	return ok && builtin.Name() == "make"
}
