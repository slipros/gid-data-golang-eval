// Package intransaction implements rule GID-175 (in-transaction):
// the transaction-handling convention.
//
// Convention (requirement 2026-06-07): transaction types live in /domain/model;
// service/usecase use them; the connection that implements the transaction
// signature is passed directly into the constructor. The canonical form in
// /domain/model:
//
//	type InTransactionFunc func(ctx context.Context, fn func(ctx context.Context) error) error
//
//	type InTransactionWithReturnFunc[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error)
//
// Checks:
//
//  1. Tx-type declaration outside model: a named func type with a tx signature
//     declared in a package OTHER than /domain/model → the transaction type lives in /domain/model.
//  2. Naming in model: in /domain/model a func type with a tx signature must be named
//     InTransactionFunc (non-generic) / InTransactionWithReturnFunc (generic).
//  3. Anonymous signature in service/usecase: in /domain/service and /domain/usecase
//     a struct field or a function/constructor parameter with an anonymous func type
//     of the tx signature → use the named type model.InTransactionFunc.
//  4. Tx-method on repo/service: in /dal/repository and /domain/service a struct
//     method with a tx signature → InTransactionFunc is passed into the constructor
//     directly from the connection; the repository/service does not wrap the transaction in a method.
//
// context.Context is recognized structurally via go/types (package context,
// name Context). The signature is matched structurally. Generated code is skipped.
package intransaction

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-175"

const (
	txNone txKind = iota
	txPlain
	txWithReturn
)

// Analyzer is rule GID-175: transaction types (InTransactionFunc) live in /domain/model.
var Analyzer = &analysis.Analyzer{
	Name: "gidintransaction",
	Doc:  ruleID + ": transaction convention; InTransactionFunc lives in /domain/model. Fix: declare it there",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	inModel := pathseg.HasLayer(pkgPath, "domain", "model")
	inServiceOrUsecase := pathseg.HasLayer(pkgPath, "domain", "service") ||
		pathseg.HasLayer(pkgPath, "domain", "usecase")
	inTxMethodScope := pathseg.HasLayer(pkgPath, "dal", "repository") ||
		pathseg.HasLayer(pkgPath, "domain", "service")

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				checkTypeDecls(pass, d, inModel)
			case *ast.FuncDecl:
				if inTxMethodScope && d.Recv != nil {
					checkTxMethod(pass, d)
				}
				if inServiceOrUsecase {
					checkAnonInParams(pass, d.Type)
				}
			}
		}
		if inServiceOrUsecase {
			checkAnonInStructFields(pass, file)
		}
	}
	return nil, nil
}

// --- Checks 1 and 2: declarations of named func types ---

func checkTypeDecls(pass *analysis.Pass, gd *ast.GenDecl, inModel bool) {
	const (
		nameInTransaction           = "InTransactionFunc"
		nameInTransactionWithReturn = "InTransactionWithReturnFunc"
	)
	for _, spec := range gd.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		obj := pass.TypesInfo.Defs[ts.Name]
		if obj == nil {
			continue
		}
		objType := obj.Type()
		sig, ok := objType.Underlying().(*types.Signature)
		if !ok {
			continue
		}
		kind := classifyTxSignature(sig)
		if kind == txNone {
			continue
		}
		if !inModel {
			// Check 1: tx-type declared outside /domain/model.
			pass.Reportf(ts.Name.Pos(),
				"%s: the transaction type must live in /domain/model (InTransactionFunc). Fix: move it there", ruleID)
			continue
		}
		// Check 2: in model the name must be canonical.
		want := nameInTransaction
		if kind == txWithReturn {
			want = nameInTransactionWithReturn
		}
		if ts.Name.Name != want {
			pass.Reportf(ts.Name.Pos(),
				"%s: the transaction type must be named InTransactionFunc / InTransactionWithReturnFunc. Fix: rename it", ruleID)
		}
	}
}

// --- Check 4: tx-method on repo/service ---

func checkTxMethod(pass *analysis.Pass, fn *ast.FuncDecl) {
	obj, ok := pass.TypesInfo.Defs[fn.Name].(*types.Func)
	if !ok {
		return
	}
	sig, ok := obj.Type().(*types.Signature)
	if !ok {
		return
	}
	if classifyTxSignature(sig) != txPlain {
		return
	}
	pass.Reportf(fn.Name.Pos(),
		"%s: a repository/service must not wrap a transaction in a method. "+
			"Fix: pass InTransactionFunc into the constructor directly from the connection", ruleID)
}

// --- Check 3: anonymous tx-signature in function/constructor parameters ---

func checkAnonInParams(pass *analysis.Pass, ft *ast.FuncType) {
	if ft.Params == nil {
		return
	}
	for _, field := range ft.Params.List {
		reportIfAnonTxField(pass, field.Type)
	}
}

// --- Check 3: anonymous tx-signature in struct fields ---

func checkAnonInStructFields(pass *analysis.Pass, file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		st, ok := n.(*ast.StructType)
		if !ok || st.Fields == nil {
			return true
		}
		for _, field := range st.Fields.List {
			reportIfAnonTxField(pass, field.Type)
		}
		return true
	})
}

// reportIfAnonTxField flags a field/parameter with an anonymous func type of the tx signature.
// A named type (including model.InTransactionFunc) is an *ast.Ident /
// *ast.SelectorExpr, not an *ast.FuncType, so it is not flagged.
func reportIfAnonTxField(pass *analysis.Pass, expr ast.Expr) {
	ftLit, ok := expr.(*ast.FuncType)
	if !ok {
		return
	}
	t := pass.TypesInfo.TypeOf(expr)
	if t == nil {
		return
	}
	sig, ok := t.Underlying().(*types.Signature)
	if !ok {
		return
	}
	if classifyTxSignature(sig) == txNone {
		return
	}
	pass.Reportf(ftLit.Pos(),
		"%s: use the named type model.InTransactionFunc. Fix: replace the anonymous signature", ruleID)
}

// --- Structural recognition of the tx-signature ---

type txKind int

// classifyTxSignature structurally matches a func type's signature against the tx forms:
//
//	plain:      func(context.Context, func(context.Context) error) error
//	withReturn: func(context.Context, func(context.Context) (T, error)) (T, error)
func classifyTxSignature(sig *types.Signature) txKind {
	params := sig.Params()
	results := sig.Results()
	if params.Len() != 2 || results.Len() == 0 {
		return txNone
	}
	param0 := params.At(0)
	if !isContextContext(param0.Type()) {
		return txNone
	}
	param1 := params.At(1)
	param1Type := param1.Type()
	cb, ok := param1Type.Underlying().(*types.Signature)
	if !ok {
		return txNone
	}
	// callback: first parameter is context.Context, exactly one parameter.
	cbParams := cb.Params()
	if cbParams.Len() != 1 {
		return txNone
	}
	cbParam0 := cbParams.At(0)
	if !isContextContext(cbParam0.Type()) {
		return txNone
	}

	cbResults := cb.Results()

	// plain: results = (error); callback results = (error).
	if results.Len() == 1 {
		result0 := results.At(0)
		if !isError(result0.Type()) {
			return txNone
		}
		if cbResults.Len() == 1 {
			cbResult0 := cbResults.At(0)
			if isError(cbResult0.Type()) {
				return txPlain
			}
		}
		return txNone
	}

	// withReturn: results = (T, error); callback results = (T, error),
	// where T matches.
	if results.Len() == 2 {
		result1 := results.At(1)
		if !isError(result1.Type()) {
			return txNone
		}
		if cbResults.Len() != 2 {
			return txNone
		}
		cbResult1 := cbResults.At(1)
		if !isError(cbResult1.Type()) {
			return txNone
		}
		result0 := results.At(0)
		cbResult0 := cbResults.At(0)
		if types.Identical(result0.Type(), cbResult0.Type()) {
			return txWithReturn
		}
	}
	return txNone
}

// isContextContext: the type is context.Context (package context, name Context).
func isContextContext(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	pkg := obj.Pkg()
	return pkg.Path() == "context" && obj.Name() == "Context"
}

func isError(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() == nil && obj.Name() == "error"
}
