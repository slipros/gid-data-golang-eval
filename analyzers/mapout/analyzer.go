// Package mapout implements rule GID-257: a map received as a PARAMETER is not
// filled in — the function returns its result.
//
// Owner's principle: what a function produces belongs in its signature. A map
// passed in as an accumulator hides the result from the type system — the
// signature says the function produces nothing, so a caller reading it cannot
// tell what comes back, and every outcome looks identical from the outside:
//
//	func (a *AdCabinetResolver) resolveChunk(ctx context.Context, chunk []uuid.UUID,
//		result map[uuid.UUID]model.AdCabinet) {
//		…
//		result[id] = model.AdCabinet{…}
//	}
//
// A chunk that resolved nothing and a chunk that failed leave the caller in
// exactly the same state (incident 2026-08-06, advertising-api
// internal/domain/service/ad_cabinet_resolver.go). Returning the map instead
// puts the result back where the reader looks for it — and leaves room for the
// error result the failing branch needs (GID-258).
//
// Detect: a function or method with a parameter of map type that the body only
// WRITES to —
//
//	param[k] = v  ·  param[k]++  ·  param[k] += 1  ·  delete(param, k)  ·  clear(param)
//
// Reading a map parameter (a lookup, a range, len) is not a write: passing a
// map IN as data is fine, passing it in to be FILLED is not.
//
// The write-only discriminator: a parameter the function also READS is state
// the caller lends it, not a result it produces, and returning it would make
// the code worse —
//
//	if visited[node] { return }; visited[node] = true   // cycle guard in a recursive walk
//	if v, ok := cache[t]; ok { return v }; cache[t] = v // memoization across calls
//
// Both are read-then-write; the incident shape (result[id] = cabinet, never
// read back) is not. Handing the parameter on to another function counts as a
// read — a recursive walk passes its visited set down, and what the callee does
// with it is not visible here.
//
// A map RECEIVER (a method on a named map type) is not a parameter and is left
// alone — mutating your own value is what such a type is for. So is a write to
// a map held in a struct field or a local variable.
//
// Generated code is skipped. A _test.go file IS judged: a fixture can return
// its map exactly as easily as production code (GID-250), and a double has no
// production interface with a fill-in-the-map method to mirror — this rule
// removes those.
package mapout

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
)

const ruleID = "GID-257"

// Analyzer — rule GID-257 with no exclusions configured.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — functions allowed to fill a map parameter:
	// "Function" (any receiver) or "Type.Method".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-257 analyzer from linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidmapout",
		Doc: ruleID + ": a map received as a parameter is filled in instead of being returned — the result " +
			"is missing from the signature. " +
			"Fix: return the map (func resolve(ctx, ids) map[K]V), do not accept it as an accumulator",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s.Exclude)
		},
	}
}

func run(pass *analysis.Pass, excluded []string) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if exclude.Match(excluded, receiverType(fn), fn.Name.Name) {
				continue
			}
			checkFunc(pass, fn)
		}
	}
	return nil, nil
}

// checkFunc reports the writes to every map parameter of fn that the function
// only WRITES to — see the write-only discriminator.
func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	params := mapParams(pass, fn)
	if len(params) == 0 {
		return
	}
	writes, readObjs := classifyUses(pass, fn.Body, params)
	for obj, at := range writes {
		if readObjs[obj] {
			continue // read as well as written: state, not a result
		}
		for _, expr := range at {
			pass.Reportf(expr.Pos(),
				"%s: a map received as a parameter is filled in — the result is missing from the signature, "+
					"so the caller cannot tell an empty result from a failure. "+
					"Fix: return it instead (func resolve(ctx, ids) map[K]V) and let the caller merge; "+
					"a map parameter is for data going IN, not for a result coming OUT",
				ruleID)
		}
	}
}

// classifyUses splits every use of a map parameter inside body into WRITES
// (the expressions to report) and READS (any other use of the identifier).
//
// The write-only discriminator: a parameter the function only writes to is an
// out-parameter — the result lives outside the signature. A parameter the
// function also READS is state the caller lends it, and returning it would
// make the code worse rather than better:
//
//	if visited[node] { return }; visited[node] = true   // cycle guard in a recursive walk
//	if v, ok := cache[t]; ok { return v }; cache[t] = v // memoization across calls
//
// Both are read-then-write; the incident shape (result[id] = cabinet, never
// read) is not. Passing the parameter on to another function counts as a read:
// a recursive walk hands its visited set down, and the analyzer cannot see what
// the callee does with it — silence is the safe answer there.
func classifyUses(
	pass *analysis.Pass, body *ast.BlockStmt, params map[types.Object]bool,
) (writes map[types.Object][]ast.Expr, reads map[types.Object]bool) {
	writes = map[types.Object][]ast.Expr{}
	reads = map[types.Object]bool{}
	writeIdents := map[*ast.Ident]bool{}

	// Pass 1: collect the writes and the identifiers that make them up, so that
	// pass 2 can tell a write's own identifier from a genuine read.
	record := func(obj types.Object, id *ast.Ident, at ast.Expr) {
		if obj == nil {
			return
		}
		writes[obj] = append(writes[obj], at)
		writeIdents[id] = true
	}
	ast.Inspect(body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range stmt.Lhs {
				record(indexWrite(pass, lhs, params))
			}
		case *ast.IncDecStmt:
			record(indexWrite(pass, stmt.X, params))
		case *ast.CallExpr:
			record(builtinWrite(pass, stmt, params))
		}
		return true
	})

	// Pass 2: every other mention of the parameter is a read.
	ast.Inspect(body, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || writeIdents[id] {
			return true
		}
		if obj := paramObject(pass, id, params); obj != nil {
			reads[obj] = true
		}
		return true
	})
	return writes, reads
}

// indexWrite reports the map parameter expr writes to when it indexes one on
// the left-hand side of an assignment (result[id] = …, result[id]++): the
// parameter's object, the identifier naming it (so the read pass can skip it),
// and the expression to report. All three are nil when expr is not such a write.
func indexWrite(
	pass *analysis.Pass, expr ast.Expr, params map[types.Object]bool,
) (obj types.Object, id *ast.Ident, at ast.Expr) {
	index, ok := expr.(*ast.IndexExpr)
	if !ok {
		return nil, nil, nil
	}
	id, ok = index.X.(*ast.Ident)
	if !ok {
		return nil, nil, nil
	}
	obj = paramObject(pass, id, params)
	if obj == nil {
		return nil, nil, nil
	}
	return obj, id, expr
}

// builtinWrite reports the map parameter mutated by delete(param, k) or
// clear(param) — the two builtins that mutate a map without an assignment.
func builtinWrite(
	pass *analysis.Pass, call *ast.CallExpr, params map[types.Object]bool,
) (obj types.Object, id *ast.Ident, at ast.Expr) {
	fun, ok := call.Fun.(*ast.Ident)
	if !ok || (fun.Name != "delete" && fun.Name != "clear") {
		return nil, nil, nil
	}
	if _, isBuiltin := pass.TypesInfo.Uses[fun].(*types.Builtin); !isBuiltin {
		return nil, nil, nil // a local function named delete/clear is not the builtin
	}
	if len(call.Args) == 0 {
		return nil, nil, nil
	}
	id, ok = call.Args[0].(*ast.Ident)
	if !ok {
		return nil, nil, nil
	}
	obj = paramObject(pass, id, params)
	if obj == nil {
		return nil, nil, nil
	}
	return obj, id, call
}

// mapParams collects the objects of fn's named parameters of map type. The
// RECEIVER is deliberately not included: a method on a named map type mutates
// its own value, which is what such a type exists for.
func mapParams(pass *analysis.Pass, fn *ast.FuncDecl) map[types.Object]bool {
	out := map[types.Object]bool{}
	if fn.Type.Params == nil {
		return out
	}
	for _, field := range fn.Type.Params.List {
		fieldType := pass.TypesInfo.TypeOf(field.Type)
		if fieldType == nil {
			continue
		}
		if _, isMap := fieldType.Underlying().(*types.Map); !isMap {
			continue
		}
		for _, name := range field.Names {
			if obj := pass.TypesInfo.Defs[name]; obj != nil {
				out[obj] = true
			}
		}
	}
	return out
}

// paramObject returns the object expr refers to when it is a plain identifier
// resolving to one of params; nil otherwise.
func paramObject(pass *analysis.Pass, expr ast.Expr, params map[types.Object]bool) types.Object {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return nil
	}
	obj := pass.TypesInfo.Uses[id]
	if obj == nil {
		obj = pass.TypesInfo.Defs[id]
	}
	if obj == nil || !params[obj] {
		return nil
	}
	return obj
}

// receiverType returns the name of fn's receiver type ("Resolver" for both
// Resolver and *Resolver), or "" for a plain function — the form
// settings.exclude matches against.
func receiverType(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	expr := fn.Recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}
