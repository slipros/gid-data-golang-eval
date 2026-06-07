// Package intransaction реализует правило GID-175 (in-transaction):
// конвенция работы с транзакциями.
//
// Конвенция (требование 2026-06-07): типы транзакций живут в /domain/model;
// service/usecase их используют; connection, реализующий сигнатуру транзакции,
// передаётся напрямую в конструктор. Каноническая форма в /domain/model:
//
//	type InTransactionFunc func(ctx context.Context, fn func(ctx context.Context) error) error
//
//	type InTransactionWithReturnFunc[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error)
//
// Проверки:
//
//  1. Объявление tx-типа вне model: именованный func-тип с tx-сигнатурой,
//     объявленный в пакете НЕ /domain/model → тип транзакции живёт в /domain/model.
//  2. Нейминг в model: в /domain/model func-тип с tx-сигнатурой обязан называться
//     InTransactionFunc (без generic) / InTransactionWithReturnFunc (generic).
//  3. Анонимная сигнатура в service/usecase: в /domain/service и /domain/usecase
//     поле структуры или параметр функции/конструктора с анонимным func-типом
//     tx-сигнатуры → используйте именованный тип model.InTransactionFunc.
//  4. Tx-метод на repo/service: в /dal/repository и /domain/service метод
//     структуры с tx-сигнатурой → InTransactionFunc передаётся в конструктор
//     напрямую от connection, репозиторий/сервис не оборачивает транзакцию методом.
//
// Распознавание context.Context — структурно через go/types (пакет context,
// имя Context). Сигнатура матчится структурно. Сгенерированный код пропускается.
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

// Analyzer — правило GID-175: типы транзакций (InTransactionFunc) живут в /domain/model.
var Analyzer = &analysis.Analyzer{
	Name: "gidintransaction",
	Doc:  ruleID + ": конвенция транзакций — InTransactionFunc живёт в /domain/model",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	inModel := pathseg.Contains(pkgPath, "domain", "model")
	inServiceOrUsecase := pathseg.Contains(pkgPath, "domain", "service") ||
		pathseg.Contains(pkgPath, "domain", "usecase")
	inTxMethodScope := pathseg.Contains(pkgPath, "dal", "repository") ||
		pathseg.Contains(pkgPath, "domain", "service")

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

// --- Проверки 1 и 2: объявления именованных func-типов ---

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
			// Проверка 1: tx-тип объявлен вне /domain/model.
			pass.Reportf(ts.Name.Pos(),
				"%s: тип транзакции живёт в /domain/model (InTransactionFunc)", ruleID)
			continue
		}
		// Проверка 2: в model имя обязано быть каноническим.
		want := nameInTransaction
		if kind == txWithReturn {
			want = nameInTransactionWithReturn
		}
		if ts.Name.Name != want {
			pass.Reportf(ts.Name.Pos(),
				"%s: тип транзакции называется InTransactionFunc / InTransactionWithReturnFunc", ruleID)
		}
	}
}

// --- Проверка 4: tx-метод на repo/service ---

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
		"%s: репозиторий/сервис не оборачивает транзакцию методом — "+
			"InTransactionFunc передаётся в конструктор напрямую от connection", ruleID)
}

// --- Проверка 3: анонимная tx-сигнатура в параметрах функций/конструкторов ---

func checkAnonInParams(pass *analysis.Pass, ft *ast.FuncType) {
	if ft.Params == nil {
		return
	}
	for _, field := range ft.Params.List {
		reportIfAnonTxField(pass, field.Type)
	}
}

// --- Проверка 3: анонимная tx-сигнатура в полях структур ---

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

// reportIfAnonTxField флагует поле/параметр с анонимным func-типом tx-сигнатуры.
// Именованный тип (в т.ч. model.InTransactionFunc) — это *ast.Ident /
// *ast.SelectorExpr, а не *ast.FuncType, поэтому он не флагуется.
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
		"%s: используйте именованный тип model.InTransactionFunc", ruleID)
}

// --- Структурное распознавание tx-сигнатуры ---

type txKind int

// classifyTxSignature структурно сопоставляет сигнатуру func-типа с tx-формами:
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
	// callback: первый параметр — context.Context, ровно один параметр.
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
	// где T совпадает.
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

// isContextContext: тип — context.Context (пакет context, имя Context).
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
