// Package dupdep implements rule GID-268: a constructor is not handed the same
// value twice under two different interfaces.
//
// The shape the rule is about:
//
//	marketplace.NewModule(
//		datalab.SavedDatasetService(),
//		datalab.ShowcaseService(),
//		datalab.ShowcaseService(), // the same entity, a second role
//	)
//
// One entity satisfies several consumer-side interfaces (GID-134 puts them next
// to the constructor), so the wiring code can pass it once per interface — and
// the constructor then holds two fields pointing at the same object, a
// dependency list that overstates how many collaborators the entity really has.
// The fix belongs to the interfaces, not to the call: merge them into one
// interface carrying both methods and take the dependency as a single parameter.
//
// FP-safe by construction. Only a constructor call is judged (New/new prefix),
// only arguments whose value is stable by inspection (a variable, a field, a
// no-argument getter — matched by the objects behind them, not by source text),
// and only when the two parameters are **different named interfaces declared in
// one non-stdlib package**: that is exactly the case where "merge them" is
// advice the caller can act on. Two parameters of the same interface type are
// left alone — there the duplicate is a wiring question, not an interface one —
// and so are stdlib interfaces (io.Copy(rw, rw) passes one value under Reader
// and Writer on purpose, and neither interface is ours to change).
package dupdep

import (
	"go/ast"
	"go/types"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-268"

// Analyzer — rule GID-268 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — settings of rule GID-268 from .golangci.yml.
type Settings struct {
	// Exclude — exclusions: "NewModule" (a constructor of any package)
	// or "marketplace.NewModule" (qualified by the calling expression).
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-268 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "giddupdep",
		Doc: ruleID + ": a constructor does not take the same dependency twice under different interfaces. " +
			"Fix: merge the interfaces into one and pass the dependency once",
		Requires: astwalk.Requires,
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	skip := func(file *ast.File) bool {
		// A test wires one double into every interface of the constructor on
		// purpose: the double is written to satisfy them all, and the fix — if
		// there is one — belongs to the production wiring the rule already sees.
		return srcfile.IsTest(pass, file) || ast.IsGenerated(file)
	}

	astwalk.NodesOf(pass, skip, func(_ *ast.File, call *ast.CallExpr) {
		checkCall(pass, s, call)
	})

	return nil, nil
}

func checkCall(pass *analysis.Pass, s Settings, call *ast.CallExpr) {
	qualifier, name, ok := calleeName(call.Fun)
	if !ok || !isCtorName(name) || exclude.Match(s.Exclude, qualifier, name) {
		return
	}

	sig, ok := pass.TypesInfo.TypeOf(call.Fun).(*types.Signature)
	if !ok {
		return
	}

	params := sig.Params()

	fixed := params.Len()
	if sig.Variadic() {
		fixed-- // the variadic tail holds no parameter of its own to name
	}
	if fixed < 2 || len(call.Args) < fixed {
		return // f(g()) spreading a multi-value call: the arguments do not line up
	}

	keys := make([]string, fixed)
	for i := range fixed {
		keys[i], _ = argKey(pass, call.Args[i])
	}

	for j := 1; j < fixed; j++ {
		if keys[j] == "" {
			continue
		}
		for i := range j {
			if keys[i] != keys[j] {
				continue
			}
			ifaceI, ifaceJ, ok := mergeableIfaces(pass, params, i, j)
			if !ok {
				continue
			}
			pass.Reportf(call.Args[j].Pos(),
				"%s: constructor %s receives the same value in parameters %s and %s — one dependency passed twice under different interfaces. "+
					"Fix: merge %s and %s into a single interface and take the dependency as one parameter",
				ruleID, name, paramDesc(params, i), paramDesc(params, j), ifaceI, ifaceJ)

			return // one diagnostic per duplicated argument
		}
	}
}

// calleeName — the name of the called function and the qualifier it is called
// through (a package name or a receiver expression), for the exclusion list.
func calleeName(fun ast.Expr) (qualifier, name string, ok bool) {
	switch e := ast.Unparen(fun).(type) {
	case *ast.Ident:
		return "", e.Name, true
	case *ast.SelectorExpr:
		if x, isIdent := ast.Unparen(e.X).(*ast.Ident); isIdent {
			return x.Name, e.Sel.Name, true
		}

		return "", e.Sel.Name, true
	case *ast.IndexExpr: // an instantiated generic constructor
		return calleeName(e.X)
	case *ast.IndexListExpr:
		return calleeName(e.X)
	}

	return "", "", false
}

// isCtorName reports whether the name is a constructor by the GID-104
// convention: the New/new prefix followed by the entity name.
func isCtorName(name string) bool {
	rest, ok := strings.CutPrefix(name, "New")
	if !ok {
		if rest, ok = strings.CutPrefix(name, "new"); !ok {
			return false
		}
	}
	if rest == "" {
		return true
	}

	r, _ := utf8.DecodeRuneInString(rest)

	return unicode.IsUpper(r)
}

// argKey identifies the value an argument stands for, or returns "" when the
// argument is not of a form whose value is stable by inspection. Matching is on
// the objects behind the expression, not on source text: dl.Showcases() and
// dl.Showcases() are the same key, s.repo and other.repo are not.
func argKey(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	switch e := ast.Unparen(expr).(type) {
	case *ast.Ident:
		obj, ok := pass.TypesInfo.ObjectOf(e).(*types.Var)
		if !ok {
			return "", false // nil, a constant, a function value
		}

		return objKey(obj), true
	case *ast.SelectorExpr:
		if pkg, isPkg := pass.TypesInfo.ObjectOf(identOf(e.X)).(*types.PkgName); isPkg {
			imported := pkg.Imported()

			return "pkg:" + imported.Path() + "." + e.Sel.Name, true
		}
		base, ok := argKey(pass, e.X)
		if !ok {
			return "", false
		}
		sel := pass.TypesInfo.ObjectOf(e.Sel)
		if sel == nil {
			return "", false
		}

		return base + "." + objKey(sel), true
	case *ast.CallExpr:
		// A getter only: an argument-taking call may well return a fresh value
		// on every call, and then the two parameters are not one dependency.
		if len(e.Args) != 0 {
			return "", false
		}

		base, ok := argKey(pass, e.Fun)
		if !ok {
			return "", false
		}

		return base + "()", true
	}

	return "", false
}

// identOf — the identifier of an expression, or nil when it is not one.
func identOf(expr ast.Expr) *ast.Ident {
	if id, ok := ast.Unparen(expr).(*ast.Ident); ok {
		return id
	}

	return nil
}

// objKey — a key identifying a types.Object across the package under analysis.
func objKey(obj types.Object) string {
	path := ""
	if pkg := obj.Pkg(); pkg != nil {
		path = pkg.Path()
	}

	return path + "." + obj.Name() + "@" + strconv.Itoa(int(obj.Pos()))
}

// mergeableIfaces reports whether parameters i and j are two different named
// interfaces from one package the code owns — the case where merging them is
// the fix. Returns their names for the diagnostic.
func mergeableIfaces(pass *analysis.Pass, params *types.Tuple, i, j int) (leftName, rightName string, ok bool) {
	leftParam, rightParam := params.At(i), params.At(j)

	left, ok := namedIface(leftParam.Type())
	if !ok {
		return "", "", false
	}

	right, ok := namedIface(rightParam.Type())
	if !ok {
		return "", "", false
	}

	leftObj, rightObj := left.Obj(), right.Obj()
	if leftObj == rightObj {
		return "", "", false // one and the same interface: nothing to merge
	}

	leftPkg, rightPkg := leftObj.Pkg(), rightObj.Pkg()
	if leftPkg != rightPkg {
		return "", "", false // declared apart: two packages, no single home to merge into
	}
	if !sameModule(pass.Pkg.Path(), leftPkg.Path()) {
		return "", "", false // io.Reader/io.Writer and the like: not ours to change
	}

	return leftObj.Name(), rightObj.Name(), true
}

// namedIface — the named type behind a parameter whose underlying type is an
// interface (the empty interface and error carry no dependency to merge).
func namedIface(t types.Type) (*types.Named, bool) {
	named, ok := t.(*types.Named)
	if !ok {
		return nil, false
	}

	obj := named.Obj()
	if obj.Pkg() == nil {
		return nil, false // error and the other universe interfaces
	}

	iface, ok := named.Underlying().(*types.Interface)
	if !ok || iface.NumMethods() == 0 {
		return nil, false
	}

	return named, true
}

// sameModule tells whether the interfaces belong to the same module as the
// package doing the wiring — the module whose interfaces the fix may merge.
// The module boundary is resolved in priority order: /internal/, then
// /pkg/<module>/, then (testdata, a non-standard layout) the first path segment.
func sameModule(pkgPath, ifacePkgPath string) bool {
	const internalSeg = "/internal/"
	if module, _, ok := strings.Cut(pkgPath, internalSeg); ok {
		return strings.HasPrefix(ifacePkgPath, module+internalSeg)
	}
	if module, ok := pathseg.PkgModuleRoot(pkgPath); ok {
		return ifacePkgPath == module || strings.HasPrefix(ifacePkgPath, module+"/")
	}

	return firstSegment(pkgPath) == firstSegment(ifacePkgPath)
}

func firstSegment(path string) string {
	seg, _, _ := strings.Cut(path, "/")

	return seg
}

// paramDesc — how a parameter is named in the diagnostic.
func paramDesc(params *types.Tuple, i int) string {
	param := params.At(i)
	pos := "#" + strconv.Itoa(i+1)
	if param.Name() == "" || param.Name() == "_" {
		return pos
	}

	return pos + " " + param.Name()
}
