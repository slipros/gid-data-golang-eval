// Package ifacemin implements rule GID-197: a dependency interface
// contains only the methods the consumer actually uses.
//
// GID-134 guarantees the interface is declared in the consumer package —
// so all uses of its methods are visible to the analyzer. A method counts as
// used when there is a reference to it (a call or a method value) outside
// *_test.go: the interface describes the needs of production code, a method
// "for the sake of a mock" is contract bloat.
//
// FP-safe: if an interface value escapes to where method consumption
// cannot be tracked (assignment/passing under a different type, a type
// assertion, a generic constraint and any unrecognized context), the
// interface is skipped entirely. Embedded interfaces are not checked.
package ifacemin

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-197"

// Analyzer — rule GID-197 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — settings of rule GID-197 from .golangci.yml.
type Settings struct {
	// Exclude — exclusions: "Interface" (as a whole) or "Interface.Method".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-197 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidifacemin",
		Doc:  ruleID + ": a dependency interface contains only the methods the consumer uses. Fix: remove unused methods from the interface",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

// ifaceDecl — a checked interface of the package.
type ifaceDecl struct {
	name     string
	typeName *types.TypeName
	methods  []*methodCand
}

type methodCand struct {
	ident *ast.Ident
	obj   types.Object
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	if !inScope(pass.Pkg.Path()) {
		return nil, nil
	}
	ifaces := collectIfaces(pass, s)
	if len(ifaces) == 0 {
		return nil, nil
	}
	used := collectUsedMethods(pass, ifaces)
	escaped := collectEscapes(pass, ifaces)
	for _, d := range ifaces {
		if escaped[d.typeName] {
			continue
		}
		for _, m := range d.methods {
			if used[m.obj] {
				continue
			}
			pass.Reportf(m.ident.Pos(),
				"%s: method %q of interface %q is not used in the consumer package. "+
					"Fix: keep the interface minimal, remove the method",
				ruleID, m.ident.Name, d.name)
		}
	}
	return nil, nil
}

// collectIfaces — the package's interfaces with their explicit methods
// (embedded interfaces are not checked).
func collectIfaces(pass *analysis.Pass, s Settings) []*ifaceDecl {
	var out []*ifaceDecl
	for _, file := range pass.Files {
		if ast.IsGenerated(file) || isTestFile(pass, file) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.TypeParams != nil {
					continue
				}
				it, ok := ts.Type.(*ast.InterfaceType)
				if !ok || slices.Contains(s.Exclude, ts.Name.Name) {
					continue
				}
				if d := newIfaceDecl(pass, s, ts, it); d != nil {
					out = append(out, d)
				}
			}
		}
	}
	return out
}

func newIfaceDecl(pass *analysis.Pass, s Settings, ts *ast.TypeSpec, it *ast.InterfaceType) *ifaceDecl {
	tn, ok := pass.TypesInfo.Defs[ts.Name].(*types.TypeName)
	if !ok {
		return nil
	}
	d := &ifaceDecl{name: ts.Name.Name, typeName: tn}
	for _, field := range it.Methods.List {
		if len(field.Names) == 0 {
			continue // an embedded interface
		}
		for _, name := range field.Names {
			if exclude.Match(s.Exclude, d.name, name.Name) {
				continue
			}
			if obj := pass.TypesInfo.Defs[name]; obj != nil {
				d.methods = append(d.methods, &methodCand{ident: name, obj: obj})
			}
		}
	}
	if len(d.methods) == 0 {
		return nil
	}
	return d
}

// collectUsedMethods — which methods of the checked interfaces are referenced
// outside *_test.go (a call and a method value; via embedding the reference
// resolves to the same method object).
func collectUsedMethods(pass *analysis.Pass, ifaces []*ifaceDecl) map[types.Object]bool {
	cands := map[types.Object]struct{}{}
	for _, d := range ifaces {
		for _, m := range d.methods {
			cands[m.obj] = struct{}{}
		}
	}
	used := map[types.Object]bool{}
	for _, file := range pass.Files {
		if isTestFile(pass, file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if !ok {
				return true
			}
			obj := pass.TypesInfo.Uses[id]
			if obj == nil {
				return true
			}
			if _, ok := cands[obj]; ok {
				used[obj] = true
			}
			return true
		})
	}
	return used
}

// collectEscapes marks interfaces whose values escape into contexts where
// method consumption cannot be tracked — those are skipped entirely.
func collectEscapes(pass *analysis.Pass, ifaces []*ifaceDecl) map[*types.TypeName]bool {
	checked := map[*types.TypeName]bool{}
	for _, d := range ifaces {
		checked[d.typeName] = true
	}
	escaped := map[*types.TypeName]bool{}
	for _, file := range pass.Files {
		if isTestFile(pass, file) {
			continue
		}
		parents := parentMap(file)
		ast.Inspect(file, func(n ast.Node) bool {
			e, ok := n.(ast.Expr)
			if !ok {
				return true
			}
			tv, ok := pass.TypesInfo.Types[e]
			if !ok {
				return true
			}
			tn := checkedIface(checked, tv.Type)
			if tn == nil || escaped[tn] {
				return true
			}
			if tv.IsValue() {
				if !safeContext(pass, parents, e) {
					escaped[tn] = true
				}
				return true
			}
			// A type in a generic constraint: calls through a type parameter
			// resolve to the constraint's object — untrackable.
			if tv.IsType() && inTypeParams(parents, e) {
				escaped[tn] = true
			}
			return true
		})
	}
	return escaped
}

func checkedIface(checked map[*types.TypeName]bool, t types.Type) *types.TypeName {
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return nil
	}
	if tn := named.Obj(); checked[tn] {
		return tn
	}
	return nil
}

// parentMap — the parent of every node in the file.
func parentMap(file *ast.File) map[ast.Node]ast.Node {
	parents := map[ast.Node]ast.Node{}
	var stack []ast.Node
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if len(stack) > 0 {
			parents[n] = stack[len(stack)-1]
		}
		stack = append(stack, n)
		return true
	})
	return parents
}

// safeContext: the interface value is used so that the consumption of its
// methods stays visible (a method call, storing/passing under the same
// type, a comparison). Any unrecognized context is not safe.
func safeContext(pass *analysis.Pass, parents map[ast.Node]ast.Node, e ast.Expr) bool {
	p := parents[e]
	for {
		pe, ok := p.(*ast.ParenExpr)
		if !ok {
			break
		}
		e = pe
		p = parents[pe]
	}
	eType := pass.TypesInfo.TypeOf(e)
	switch ctx := p.(type) {
	case *ast.SelectorExpr:
		// e.M — a call/method value (covered by the Uses scan); e == Sel —
		// a field value, classified at the level of the selector's parent.
		return true
	case *ast.CallExpr:
		if ctx.Fun == e {
			return true
		}
		return argIdentical(pass, ctx, e, eType)
	case *ast.AssignStmt:
		return assignIdentical(pass, ctx, e, eType)
	case *ast.ValueSpec:
		return valueSpecIdentical(pass, ctx, e, eType)
	case *ast.KeyValueExpr:
		return keyValueIdentical(pass, parents, ctx, e, eType)
	case *ast.CompositeLit:
		return elemIdentical(pass, ctx, e, eType)
	case *ast.ReturnStmt:
		return returnIdentical(pass, parents, ctx, e, eType)
	case *ast.BinaryExpr, *ast.ExprStmt, *ast.CaseClause, *ast.SwitchStmt:
		return true // a comparison or a bare expression — no methods consumed
	default:
		return false
	}
}

// argIdentical: the argument lands in a parameter of the same interface type.
func argIdentical(pass *analysis.Pass, call *ast.CallExpr, e ast.Expr, eType types.Type) bool {
	funType := pass.TypesInfo.TypeOf(call.Fun)
	if funType == nil {
		return false
	}
	if tv, ok := pass.TypesInfo.Types[call.Fun]; ok && tv.IsType() {
		return types.Identical(funType, eType) // a conversion
	}
	sig, ok := funType.Underlying().(*types.Signature)
	if !ok {
		return false // a builtin, etc.
	}
	idx := slices.IndexFunc(call.Args, func(a ast.Expr) bool { return a == e })
	if idx < 0 {
		return false
	}
	params := sig.Params()
	switch {
	case sig.Variadic() && idx >= params.Len()-1:
		lastParam := params.At(params.Len() - 1)
		last := lastParam.Type()
		if sl, ok := last.(*types.Slice); ok && !call.Ellipsis.IsValid() {
			return types.Identical(sl.Elem(), eType)
		}
		return types.Identical(last, eType)
	case idx < params.Len():
		param := params.At(idx)
		return types.Identical(param.Type(), eType)
	default:
		return false
	}
}

func assignIdentical(pass *analysis.Pass, st *ast.AssignStmt, e ast.Expr, eType types.Type) bool {
	if len(st.Lhs) != len(st.Rhs) {
		return false
	}
	idx := slices.IndexFunc(st.Rhs, func(a ast.Expr) bool { return a == e })
	if idx < 0 {
		return true // e is in Lhs — a write into it, no methods consumed
	}
	if id, ok := st.Lhs[idx].(*ast.Ident); ok && id.Name == "_" {
		return true
	}
	lt := pass.TypesInfo.TypeOf(st.Lhs[idx])
	return lt != nil && types.Identical(lt, eType)
}

func valueSpecIdentical(pass *analysis.Pass, vs *ast.ValueSpec, e ast.Expr, eType types.Type) bool {
	if len(vs.Names) != len(vs.Values) {
		return false
	}
	idx := slices.IndexFunc(vs.Values, func(a ast.Expr) bool { return a == e })
	if idx < 0 {
		return false
	}
	if vs.Names[idx].Name == "_" {
		return true
	}
	obj := pass.TypesInfo.Defs[vs.Names[idx]]
	return obj != nil && types.Identical(obj.Type(), eType)
}

func keyValueIdentical(
	pass *analysis.Pass,
	parents map[ast.Node]ast.Node,
	kv *ast.KeyValueExpr,
	e ast.Expr,
	eType types.Type,
) bool {
	lit, ok := parents[kv].(*ast.CompositeLit)
	if !ok {
		return false
	}
	lt := pass.TypesInfo.TypeOf(lit)
	if lt == nil {
		return false
	}
	switch u := lt.Underlying().(type) {
	case *types.Struct:
		key, ok := kv.Key.(*ast.Ident)
		if !ok || e != kv.Value {
			return false
		}
		for f := range u.Fields() {
			if f.Name() == key.Name {
				return types.Identical(f.Type(), eType)
			}
		}
		return false
	case *types.Map:
		if e == kv.Key {
			return types.Identical(u.Key(), eType)
		}
		return types.Identical(u.Elem(), eType)
	default:
		return false
	}
}

func elemIdentical(pass *analysis.Pass, lit *ast.CompositeLit, e ast.Expr, eType types.Type) bool {
	lt := pass.TypesInfo.TypeOf(lit)
	if lt == nil {
		return false
	}
	switch u := lt.Underlying().(type) {
	case *types.Slice:
		return types.Identical(u.Elem(), eType)
	case *types.Array:
		return types.Identical(u.Elem(), eType)
	case *types.Struct:
		idx := slices.IndexFunc(lit.Elts, func(a ast.Expr) bool { return a == e })
		if idx < 0 || idx >= u.NumFields() {
			return false
		}
		field := u.Field(idx)
		return types.Identical(field.Type(), eType)
	default:
		return false
	}
}

func returnIdentical(
	pass *analysis.Pass,
	parents map[ast.Node]ast.Node,
	ret *ast.ReturnStmt,
	e ast.Expr,
	eType types.Type,
) bool {
	sig := enclosingSignature(pass, parents, ret)
	if sig == nil {
		return false
	}
	results := sig.Results()
	if len(ret.Results) != results.Len() {
		return false
	}
	idx := slices.IndexFunc(ret.Results, func(a ast.Expr) bool { return a == e })
	if idx < 0 {
		return false
	}
	result := results.At(idx)
	return types.Identical(result.Type(), eType)
}

func enclosingSignature(pass *analysis.Pass, parents map[ast.Node]ast.Node, n ast.Node) *types.Signature {
	for cur := parents[n]; cur != nil; cur = parents[cur] {
		switch fn := cur.(type) {
		case *ast.FuncLit:
			if sig, ok := pass.TypesInfo.TypeOf(fn).(*types.Signature); ok {
				return sig
			}
			return nil
		case *ast.FuncDecl:
			obj := pass.TypesInfo.Defs[fn.Name]
			if obj == nil {
				return nil
			}
			if sig, ok := obj.Type().(*types.Signature); ok {
				return sig
			}
			return nil
		}
	}
	return nil
}

// inTypeParams: the type is used in a type parameter list (a constraint).
func inTypeParams(parents map[ast.Node]ast.Node, e ast.Expr) bool {
	for cur := parents[e]; cur != nil; cur = parents[cur] {
		switch p := cur.(type) {
		case *ast.FuncDecl:
			return false
		case *ast.TypeSpec:
			return false
		case *ast.FieldList:
			if ft, ok := parents[p].(*ast.FuncType); ok && ft.TypeParams == p {
				return true
			}
			if ts, ok := parents[p].(*ast.TypeSpec); ok && ts.TypeParams == p {
				return true
			}
		}
	}
	return false
}

func inScope(pkgPath string) bool {
	return pathseg.Contains(pkgPath, "domain", "service") ||
		pathseg.Contains(pkgPath, "domain", "usecase") ||
		pathseg.Contains(pkgPath, "dal", "repository") ||
		pathseg.Contains(pkgPath, "server") ||
		pathseg.Contains(pkgPath, "event")
}

func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	return strings.HasSuffix(filepath.Base(pass.Fset.Position(file.Pos()).Filename), "_test.go")
}
