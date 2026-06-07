// Package patterns implements the simple AST-pattern rules that used to live
// in the ruleguard layer (ruleguard/rules.go). They are now native
// go/analysis analyzers so the rules ship inside the binary — consumers no
// longer copy a rules file next to their config.
//
// One rule = one linter (independently togglable in .golangci.yml):
//   - GID-001 gidtimenow     — no direct time.Now()
//   - GID-002 giduuidnil      — compare UUID via IsNil(), not uuid.UUID{}
//   - GID-003 giduuidversion  — generate UUIDs via uuid.NewV7()
//   - GID-005 gidnewderef     — avoid the new() builtin
//   - GID-006 gidyoda         — no yoda conditions
//   - GID-007 gidquoteverb    — use %q instead of hand-escaped quotes
//   - GID-008 giddeepequal    — avoid reflect.DeepEqual
package patterns

import (
	"bytes"
	"go/ast"
	"go/constant"
	"go/printer"
	"go/token"
	"go/types"
	"regexp"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const gofrsUUID = "github.com/gofrs/uuid"

// TimeNowAnalyzer — GID-001.
var TimeNowAnalyzer = &analysis.Analyzer{
	Name: "gidtimenow",
	Doc:  "GID-001: time.Now() must not be called directly. Fix: use gdhelper.StdTime.Now() instead of time.Now().",
	Run:  runTimeNow,
}

// UUIDNilAnalyzer — GID-002.
var UUIDNilAnalyzer = &analysis.Analyzer{
	Name: "giduuidnil",
	Doc:  `GID-002: do not compare a UUID with uuid.UUID{}. Fix: replace "id == uuid.UUID{}" with "id.IsNil()".`,
	Run:  runUUIDNil,
}

// UUIDVersionAnalyzer — GID-003.
var UUIDVersionAnalyzer = &analysis.Analyzer{
	Name: "giduuidversion",
	Doc:  "GID-003: UUIDs must be generated uniformly. Fix: use uuid.Must(uuid.NewV7()) instead of uuid.NewV1/3/4/5/6().",
	Run:  runUUIDVersion,
}

// NewDerefAnalyzer — GID-005.
var NewDerefAnalyzer = &analysis.Analyzer{
	Name: "gidnewderef",
	Doc:  `GID-005: avoid the new() builtin. Fix: use "&T{}" for structs or "var x T" instead of "new(T)".`,
	Run:  runNewDeref,
}

// YodaAnalyzer — GID-006.
var YodaAnalyzer = &analysis.Analyzer{
	Name: "gidyoda",
	Doc:  `GID-006: yoda condition — the literal must be on the right. Fix: write "x == 0" instead of "0 == x".`,
	Run:  runYoda,
}

// QuoteVerbAnalyzer — GID-007.
var QuoteVerbAnalyzer = &analysis.Analyzer{
	Name: "gidquoteverb",
	Doc:  `GID-007: do not escape quotes around %s/%v by hand. Fix: use %q instead of \"%s\".`,
	Run:  runQuoteVerb,
}

// DeepEqualAnalyzer — GID-008.
var DeepEqualAnalyzer = &analysis.Analyzer{
	Name: "giddeepequal",
	Doc:  "GID-008: avoid reflect.DeepEqual. Fix: use require/cmp in tests or explicit field comparison in code.",
	Run:  runDeepEqual,
}

// uuidVersionFuncs — generator functions banned in favour of NewV7.
var uuidVersionFuncs = map[string]struct{}{
	"NewV1": {}, "NewV3": {}, "NewV4": {}, "NewV5": {}, "NewV6": {},
}

// printfFormatArg maps a printf-like function (by package path + name) to the
// index of its format-string argument.
var printfFormatArg = map[printfFunc]int{
	{"fmt", "Sprintf"}:                  0,
	{"fmt", "Errorf"}:                   0,
	{"fmt", "Printf"}:                   0,
	{"fmt", "Fprintf"}:                  1,
	{"github.com/pkg/errors", "Errorf"}: 0,
	{"github.com/pkg/errors", "Wrapf"}:  1,
}

// handEscapedQuote matches a hand-escaped quoted verb: "%s" or "%v" with
// literal double quotes around the verb.
var handEscapedQuote = regexp.MustCompile(`"%[sv]"`)

// printfFunc identifies a printf-like function by its package path and name.
type printfFunc struct {
	pkg, name string
}

// funcCallee returns the function called by call, or nil if it is not a
// static function call (e.g. a builtin or a call through a variable).
func funcCallee(pass *analysis.Pass, call *ast.CallExpr) *types.Func {
	fn, ok := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
	if !ok {
		return nil
	}
	return fn
}

// calleePkgPath returns (package path, function name) of a static call, or
// ("", "") when it is not a static function from a named package.
func calleePkgPath(pass *analysis.Pass, call *ast.CallExpr) (pkgPath, funcName string) {
	fn := funcCallee(pass, call)
	if fn == nil {
		return "", ""
	}
	pkg := fn.Pkg()
	if pkg == nil {
		return "", ""
	}
	return pkg.Path(), fn.Name()
}

// inspectCalls walks every non-generated file calling fn for each CallExpr.
func inspectCalls(pass *analysis.Pass, fn func(*ast.CallExpr)) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				fn(call)
			}
			return true
		})
	}
}

// inspectBinary walks every non-generated file calling fn for each BinaryExpr.
func inspectBinary(pass *analysis.Pass, fn func(*ast.BinaryExpr)) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			if be, ok := n.(*ast.BinaryExpr); ok {
				fn(be)
			}
			return true
		})
	}
}

func runTimeNow(pass *analysis.Pass) (any, error) {
	inspectCalls(pass, func(call *ast.CallExpr) {
		pkg, name := calleePkgPath(pass, call)
		if pkg == "time" && name == "Now" {
			pass.Reportf(call.Pos(),
				"GID-001: time.Now() must not be called directly. "+
					"Fix: use gdhelper.StdTime.Now() instead of time.Now().")
		}
	})
	return nil, nil
}

func runDeepEqual(pass *analysis.Pass) (any, error) {
	inspectCalls(pass, func(call *ast.CallExpr) {
		pkg, name := calleePkgPath(pass, call)
		if pkg == "reflect" && name == "DeepEqual" {
			pass.Reportf(call.Pos(),
				"GID-008: avoid reflect.DeepEqual. "+
					"Fix: use require/cmp in tests or explicit field comparison in code.")
		}
	})
	return nil, nil
}

func runUUIDVersion(pass *analysis.Pass) (any, error) {
	inspectCalls(pass, func(call *ast.CallExpr) {
		pkg, name := calleePkgPath(pass, call)
		if pkg != gofrsUUID {
			return
		}
		if _, ok := uuidVersionFuncs[name]; ok {
			pass.Reportf(call.Pos(),
				"GID-003: UUIDs must be generated uniformly. "+
					"Fix: use uuid.Must(uuid.NewV7()) instead of uuid.%s().", name)
		}
	})
	return nil, nil
}

func runNewDeref(pass *analysis.Pass) (any, error) {
	inspectCalls(pass, func(call *ast.CallExpr) {
		ident, ok := call.Fun.(*ast.Ident)
		if !ok || ident.Name != "new" {
			return
		}
		if _, ok := pass.TypesInfo.ObjectOf(ident).(*types.Builtin); !ok {
			return
		}
		pass.Reportf(call.Pos(),
			`GID-005: avoid the new() builtin. `+
				`Fix: use "&T{}" for structs or "var x T" instead of "new(T)".`)
	})
	return nil, nil
}

func runYoda(pass *analysis.Pass) (any, error) {
	inspectBinary(pass, func(be *ast.BinaryExpr) {
		if be.Op != token.EQL && be.Op != token.NEQ {
			return
		}
		// Literal on the left, non-constant on the right — the yoda shape.
		if isConst(pass, be.X) && !isConst(pass, be.Y) {
			pass.Reportf(be.Pos(),
				`GID-006: yoda condition — the literal must be on the right. `+
					`Fix: write "x == 0" instead of "0 == x".`)
		}
	})
	return nil, nil
}

func runUUIDNil(pass *analysis.Pass) (any, error) {
	inspectBinary(pass, func(be *ast.BinaryExpr) {
		if be.Op != token.EQL && be.Op != token.NEQ {
			return
		}
		// One side must be uuid.UUID{}, the other a uuid.UUID value.
		var value ast.Expr
		switch {
		case uuidEmptyLit(pass, be.Y) && isGofrsUUID(pass.TypesInfo.TypeOf(be.X)):
			value = be.X
		case uuidEmptyLit(pass, be.X) && isGofrsUUID(pass.TypesInfo.TypeOf(be.Y)):
			value = be.Y
		default:
			return
		}

		valueText := render(pass.Fset, value)
		replacement := valueText + ".IsNil()"
		op := "=="
		if be.Op == token.NEQ {
			replacement = "!" + replacement
			op = "!="
		}
		pass.Report(analysis.Diagnostic{
			Pos: be.Pos(),
			Message: "GID-002: do not compare a UUID with uuid.UUID{}. " +
				`Fix: replace "` + valueText + " " + op + ` uuid.UUID{}" with "` + replacement + `".`,
			SuggestedFixes: []analysis.SuggestedFix{{
				Message: "use IsNil()",
				TextEdits: []analysis.TextEdit{{
					Pos:     be.Pos(),
					End:     be.End(),
					NewText: []byte(replacement),
				}},
			}},
		})
	})
	return nil, nil
}

func runQuoteVerb(pass *analysis.Pass) (any, error) {
	inspectCalls(pass, func(call *ast.CallExpr) {
		pkg, name := calleePkgPath(pass, call)
		idx, ok := printfFormatArg[printfFunc{pkg, name}]
		if !ok || idx >= len(call.Args) {
			return
		}
		tv, ok := pass.TypesInfo.Types[call.Args[idx]]
		if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
			return
		}
		if handEscapedQuote.MatchString(constant.StringVal(tv.Value)) {
			pass.Reportf(call.Args[idx].Pos(),
				`GID-007: do not escape quotes around %%s/%%v by hand. `+
					`Fix: use %%q instead of \"%%s\".`)
		}
	})
	return nil, nil
}

// isConst reports whether e is a constant expression.
func isConst(pass *analysis.Pass, e ast.Expr) bool {
	tv, ok := pass.TypesInfo.Types[e]
	return ok && tv.Value != nil
}

// uuidEmptyLit reports whether e is an empty composite literal of the gofrs
// uuid.UUID type (uuid.UUID{}).
func uuidEmptyLit(pass *analysis.Pass, e ast.Expr) bool {
	cl, ok := e.(*ast.CompositeLit)
	if !ok || len(cl.Elts) != 0 {
		return false
	}
	return isGofrsUUID(pass.TypesInfo.TypeOf(cl))
}

// isGofrsUUID reports whether t is github.com/gofrs/uuid.UUID.
func isGofrsUUID(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	return obj.Name() == "UUID" && pkg != nil && pkg.Path() == gofrsUUID
}

// render prints an AST expression back to source text.
func render(fset *token.FileSet, x ast.Expr) string {
	var b bytes.Buffer
	if err := printer.Fprint(&b, fset, x); err != nil {
		return ""
	}
	return b.String()
}
