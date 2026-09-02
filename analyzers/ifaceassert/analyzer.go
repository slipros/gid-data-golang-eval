// Package ifaceassert implements rule GID-274: a package-level compile-time
// interface assertion is redundant when the package itself already hands the
// concrete type over as that interface.
//
// The shape the rule is about:
//
//	// Compile-time contract check: the Iceberg read repository implements the
//	// resolver interface.
//	var _ service.DatasetSnapshotRepository = (*repository.DatasetSnapshot)(nil)
//
//	// ... and a few lines below, in the same package:
//	snapshot := service.NewDatasetSnapshot(datasetSnapshotRepo)
//
// The call is the check. Passing *repository.DatasetSnapshot into a parameter
// of type service.DatasetSnapshotRepository makes the compiler verify the very
// same contract, at the very same build, with the very same error — so the
// assertion above states a fact the line below already enforces, and every
// change to the interface now has to be carried in two places. The comment on
// top usually promises more than the assertion delivers (that the resolver goes
// through the repository and not through the client is a question about
// imports, not about method sets — GID-267 and friends judge that).
//
// The assertion is left alone wherever it does earn its place: nothing in the
// package converts the type (a library handing an implementation to consumers
// outside the module), or the wiring goes through reflection (a DI container,
// where a mismatch surfaces at runtime and this line pulls it back to compile
// time). That is exactly what "no conversion found in this package" means, so
// the rule is silent by default and speaks only on the duplicate.
//
// FP-safe by construction: the rule reports only when it has found the
// conversion that makes the assertion redundant, and it names it in the
// diagnostic. Every context it does not understand is a conversion it does not
// see, which costs a diagnostic, never a false one. An assertion in a
// production file is only ever proven by a conversion in a production file: a
// conversion living in _test.go is checked by `go test`, not by `go build`.
package ifaceassert

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strconv"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-274"

// Analyzer — rule GID-274 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — settings of rule GID-274 from .golangci.yml.
type Settings struct {
	// Exclude — exclusions: "DatasetSnapshotRepository" (an assertion of that
	// interface, whatever the concrete type) or
	// "DatasetSnapshot.DatasetSnapshotRepository" (qualified by the type the
	// assertion is written for).
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-274 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidifaceassert",
		Doc: ruleID + ": a compile-time interface assertion duplicating a conversion the same package already performs. " +
			"Fix: delete the var _ Iface = (*Type)(nil) line — the call it is wired into checks the contract",
		Requires: astwalk.Requires,
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

// assertion — one `var _ Iface = value` declaration under judgement.
type assertion struct {
	iface    types.Type // the interface being asserted
	concrete types.Type // the type the assertion is written for
	ifaceStr string     // how both are spelled in the diagnostic
	valueStr string
	pos      token.Pos
	inTest   bool
	proof    token.Pos // the conversion that makes the assertion redundant
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	found := collect(pass, s)
	if len(found) == 0 {
		return nil, nil // the package holds no assertion: nothing to walk for
	}

	findProofs(pass, found)

	for _, a := range found {
		if !a.proof.IsValid() {
			continue
		}

		pass.Reportf(a.pos,
			"%s: redundant compile-time assertion: the package already passes this value as %s at %s, "+
				"so the compiler checks the contract there. Fix: delete the \"var _ %s = %s\" line",
			ruleID, a.ifaceStr, shortPos(pass, a.proof), a.ifaceStr, a.valueStr)
	}

	return nil, nil
}

// collect gathers the package-level blank assertions of the package.
func collect(pass *analysis.Pass, s Settings) []*assertion {
	var found []*assertion

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}

		inTest := srcfile.IsTest(pass, file)

		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}

			for _, spec := range gen.Specs {
				vs, isValue := spec.(*ast.ValueSpec)
				if !isValue {
					continue
				}

				if a := assertionOf(pass, s, file, vs, inTest); a != nil {
					found = append(found, a)
				}
			}
		}
	}

	return found
}

// assertionOf recognises `var _ Iface = value` and returns it as a candidate.
func assertionOf(pass *analysis.Pass, s Settings, file *ast.File, vs *ast.ValueSpec, inTest bool) *assertion {
	if len(vs.Names) != 1 || vs.Names[0].Name != "_" || vs.Type == nil || len(vs.Values) != 1 {
		return nil
	}

	iface := pass.TypesInfo.TypeOf(vs.Type)
	if iface == nil {
		return nil
	}

	if it, ok := iface.Underlying().(*types.Interface); !ok || it.NumMethods() == 0 {
		return nil // `var _ any = x` asserts no contract to begin with
	}

	concrete := pass.TypesInfo.TypeOf(vs.Values[0])
	if concrete == nil || concrete == types.Typ[types.Invalid] {
		return nil
	}

	if _, isIface := concrete.Underlying().(*types.Interface); isIface {
		return nil // one interface asserted against another: not the wiring shape
	}

	if exclude.Match(s.Exclude, baseName(concrete), baseName(iface)) {
		return nil
	}

	return &assertion{
		iface:    iface,
		concrete: concrete,
		ifaceStr: types.TypeString(iface, qualifier(pass, file)),
		valueStr: types.ExprString(vs.Values[0]),
		pos:      vs.Pos(),
		inTest:   inTest,
	}
}

// findProofs walks the package for a conversion of the asserted type to the
// asserted interface — an argument of a call, an assignment, a declaration with
// an explicit type, a field of a composite literal, a returned value. Each is a
// place where the compiler performs the check the assertion writes out by hand.
func findProofs(pass *analysis.Pass, found []*assertion) {
	filter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
		(*ast.ReturnStmt)(nil),
		(*ast.CallExpr)(nil),
		(*ast.AssignStmt)(nil),
		(*ast.ValueSpec)(nil),
		(*ast.CompositeLit)(nil),
	}

	// results — the result tuples of the functions the walk is currently inside,
	// so that a return statement knows what it is converting to.
	var results []*types.Tuple

	var (
		curFile *ast.File
		inTest  bool
	)

	record := func(value ast.Expr, target types.Type) {
		recordProof(pass, found, value, target, inTest)
	}

	astwalk.Around(pass, filter, nil, func(file *ast.File, n ast.Node, push bool) bool {
		if file != curFile {
			curFile, inTest = file, srcfile.IsTest(pass, file)
		}

		switch node := n.(type) {
		case *ast.FuncDecl:
			results = pushPop(results, push, declResults(pass, node))
		case *ast.FuncLit:
			results = pushPop(results, push, litResults(pass, node))
		case *ast.ReturnStmt:
			if push {
				visitReturn(node, results, record)
			}
		case *ast.CallExpr:
			if push {
				visitCall(pass, node, record)
			}
		case *ast.AssignStmt:
			if push {
				visitAssign(pass, node, record)
			}
		case *ast.ValueSpec:
			if push {
				visitValueSpec(pass, node, record)
			}
		case *ast.CompositeLit:
			if push {
				visitCompositeLit(pass, node, record)
			}
		}

		return true
	})
}

// recordProof marks every assertion the conversion of value to target makes
// redundant. The interfaces must be identical: a value handed over as a wider
// interface does imply the narrower one, but the diagnostic would then name a
// contract the reader cannot find in the source, so the rule keeps quiet. On
// the concrete side a value type proves the assertion written for its pointer
// (the method set of *T contains the method set of T), never the other way
// round. An interface value handed on is not filtered out here: no assertion
// is ever recorded for one, so it matches nothing.
func recordProof(pass *analysis.Pass, found []*assertion, value ast.Expr, target types.Type, inTest bool) {
	if value == nil || target == nil {
		return
	}

	if it, ok := target.Underlying().(*types.Interface); !ok || it.NumMethods() == 0 {
		return
	}

	concrete := pass.TypesInfo.TypeOf(value)
	if concrete == nil {
		return
	}

	for _, a := range found {
		if a.proof.IsValid() || (inTest && !a.inTest) {
			continue // a conversion in _test.go is checked by go test, not go build
		}

		if types.Identical(a.iface, target) && provesConcrete(a.concrete, concrete) {
			a.proof = value.Pos()
		}
	}
}

// provesConcrete reports whether a conversion of used implies the assertion
// written for asserted.
func provesConcrete(asserted, used types.Type) bool {
	if types.Identical(asserted, used) {
		return true
	}

	ptr, ok := asserted.(*types.Pointer)

	return ok && types.Identical(ptr.Elem(), used)
}

func visitReturn(node *ast.ReturnStmt, results []*types.Tuple, record func(ast.Expr, types.Type)) {
	if len(results) == 0 {
		return
	}

	res := results[len(results)-1]
	if res == nil || len(node.Results) != res.Len() {
		return // a naked return, or one spreading a multi-value call
	}

	for i, r := range node.Results {
		result := res.At(i)

		record(r, result.Type())
	}
}

func visitCall(pass *analysis.Pass, call *ast.CallExpr, record func(ast.Expr, types.Type)) {
	sig, ok := pass.TypesInfo.TypeOf(call.Fun).(*types.Signature)
	if !ok {
		return // a conversion T(x), a builtin: no parameters to convert to
	}

	params := sig.Params()
	fixed := params.Len()

	if sig.Variadic() {
		fixed--
	}

	for i, arg := range call.Args {
		switch {
		case i < fixed:
			param := params.At(i)

			record(arg, param.Type())
		case !sig.Variadic() || call.Ellipsis.IsValid():
			return // f(g()) spreading a multi-value call, or f(xs...)
		default:
			tail := params.At(params.Len() - 1)

			slice, isSlice := tail.Type().(*types.Slice)
			if !isSlice {
				return
			}

			record(arg, slice.Elem())
		}
	}
}

func visitAssign(pass *analysis.Pass, node *ast.AssignStmt, record func(ast.Expr, types.Type)) {
	if node.Tok != token.ASSIGN || len(node.Lhs) != len(node.Rhs) {
		return // := takes the type of the value: no conversion happens
	}

	for i, lhs := range node.Lhs {
		record(node.Rhs[i], pass.TypesInfo.TypeOf(lhs))
	}
}

func visitValueSpec(pass *analysis.Pass, vs *ast.ValueSpec, record func(ast.Expr, types.Type)) {
	if vs.Type == nil || len(vs.Values) == 0 {
		return
	}

	for _, name := range vs.Names {
		if name.Name == "_" {
			return // the assertion under judgement, not a proof of it
		}
	}

	target := pass.TypesInfo.TypeOf(vs.Type)
	for _, v := range vs.Values {
		record(v, target)
	}
}

func visitCompositeLit(pass *analysis.Pass, lit *ast.CompositeLit, record func(ast.Expr, types.Type)) {
	typ := pass.TypesInfo.TypeOf(lit)
	if typ == nil {
		return
	}

	switch u := typ.Underlying().(type) {
	case *types.Struct:
		recordStructLit(u, lit, record)
	case *types.Slice:
		recordElems(lit, u.Elem(), record)
	case *types.Array:
		recordElems(lit, u.Elem(), record)
	case *types.Map:
		recordElems(lit, u.Elem(), record)
	}
}

func recordStructLit(st *types.Struct, lit *ast.CompositeLit, record func(ast.Expr, types.Type)) {
	for i, elt := range lit.Elts {
		kv, keyed := elt.(*ast.KeyValueExpr)
		if !keyed {
			if i < st.NumFields() {
				field := st.Field(i)

				record(elt, field.Type())
			}

			continue
		}

		key, isIdent := kv.Key.(*ast.Ident)
		if !isIdent {
			continue
		}

		for f := range st.NumFields() {
			field := st.Field(f)
			if field.Name() != key.Name {
				continue
			}

			record(kv.Value, field.Type())

			break
		}
	}
}

func recordElems(lit *ast.CompositeLit, elem types.Type, record func(ast.Expr, types.Type)) {
	for _, elt := range lit.Elts {
		if kv, keyed := elt.(*ast.KeyValueExpr); keyed {
			record(kv.Value, elem)

			continue
		}

		record(elt, elem)
	}
}

func declResults(pass *analysis.Pass, node *ast.FuncDecl) *types.Tuple {
	fn, ok := pass.TypesInfo.Defs[node.Name].(*types.Func)
	if !ok {
		return nil
	}

	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return nil
	}

	return sig.Results()
}

func litResults(pass *analysis.Pass, lit *ast.FuncLit) *types.Tuple {
	sig, ok := pass.TypesInfo.TypeOf(lit).(*types.Signature)
	if !ok {
		return nil
	}

	return sig.Results()
}

// pushPop keeps the stack of enclosing function results in step with the walk.
func pushPop(stack []*types.Tuple, push bool, res *types.Tuple) []*types.Tuple {
	if push {
		return append(stack, res)
	}

	if len(stack) == 0 {
		return stack
	}

	return stack[:len(stack)-1]
}

// qualifier spells a type of another package the way the file under judgement
// spells it: by the local name of the import (the alias when there is one), so
// that the interface in the diagnostic can be found in the source as written.
func qualifier(pass *analysis.Pass, file *ast.File) types.Qualifier {
	aliases := make(map[string]string, len(file.Imports))

	for _, spec := range file.Imports {
		if spec.Name == nil || spec.Name.Name == "_" || spec.Name.Name == "." {
			continue
		}

		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}

		aliases[path] = spec.Name.Name
	}

	return func(other *types.Package) string {
		if other == pass.Pkg {
			return ""
		}

		if alias, ok := aliases[other.Path()]; ok {
			return alias
		}

		return other.Name()
	}
}

// baseName — the type name without the package qualifier, for the exclusion list.
func baseName(t types.Type) string {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	named, ok := t.(*types.Named)
	if !ok {
		return ""
	}

	obj := named.Obj()

	return obj.Name()
}

// shortPos — the position of the proving conversion as file:line, so that the
// diagnostic points at the line that already performs the check.
func shortPos(pass *analysis.Pass, pos token.Pos) string {
	p := pass.Fset.Position(pos)

	return filepath.Base(p.Filename) + ":" + strconv.Itoa(p.Line)
}
