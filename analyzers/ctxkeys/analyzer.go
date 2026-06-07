// Package ctxkeys реализует правила работы с context:
//
//   - GID-165: helper'ы, складывающие данные в context, живут только в
//     /domain/model. Свой contextKey в http middleware (или любом другом
//     слое) запрещён — иначе бизнес-слои зависят от слоя middleware.
//   - GID-166: форма helper'ов в model — функция, складывающая в ctx,
//     публична и именуется ContextWith<Name>; достающая из ctx —
//     <Name>FromContext; helper живёт в одном файле с сущностью <Name>.
//   - GID-167: ключ контекста — публичный тип ContextKey
//     (type ContextKey string); const-значения ContextKey — string
//     в формате snake_case.
package ctxkeys

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const (
	rulePlace  = "GID-165"
	ruleNaming = "GID-166"
	ruleKey    = "GID-167"

	keyTypeName = "ContextKey"
)

var snakeCase = regexp.MustCompile(`^[a-z0-9]+(_[a-z0-9]+)*$`)

// Analyzer — правило GID: см. Doc.
var Analyzer = &analysis.Analyzer{
	Name: "gidctxkeys",
	Doc: rulePlace + "/" + ruleNaming + "/" + ruleKey +
		": context keys live in /domain/model; ContextWith<Name> / <Name>FromContext; type ContextKey string. Fix: move keys and helpers into /domain/model",
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	if pathseg.Contains(pass.Pkg.Path(), "domain", "model") {
		checkModelHelpers(pass)
		return nil, nil
	}
	checkNoWithValue(pass)
	return nil, nil
}

// checkNoWithValue — GID-165: вне model складывать в ctx запрещено.
func checkNoWithValue(pass *analysis.Pass) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if ok && isWithValue(pass, call) {
				pass.Reportf(call.Pos(),
					"%s: context.WithValue outside /domain/model is forbidden. Fix: keep context keys and helpers "+
						"in /domain/model so business layers do not depend on middleware",
					rulePlace)
			}
			return true
		})
	}
}

// checkModelHelpers — GID-166/167: форма helper'ов и ключей в model.
func checkModelHelpers(pass *analysis.Pass) {
	structFile := structFiles(pass)
	keyTypeFile := contextKeyFile(pass)
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Body == nil {
					continue
				}
				checkHelper(pass, file, d, structFile)
			case *ast.GenDecl:
				checkKeyConsts(pass, d, file, keyTypeFile)
			}
		}
		checkKeyTypes(pass, file)
	}
}

// contextKeyFile — файл объявления типа ContextKey.
func contextKeyFile(pass *analysis.Pass) *ast.File {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.Name == keyTypeName {
					return file
				}
			}
		}
	}
	return nil
}

// checkKeyTypes — GID-167: ключ в WithValue — публичный typed-string ContextKey.
func checkKeyTypes(pass *analysis.Pass, file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isWithValue(pass, call) || len(call.Args) != 3 {
			return true
		}
		keyType := pass.TypesInfo.TypeOf(call.Args[1])
		named, ok := keyType.(*types.Named)
		if !ok {
			pass.Reportf(call.Args[1].Pos(),
				"%s: context key must be the public type %s (type %s string), not a raw value",
				ruleKey, keyTypeName, keyTypeName)
			return true
		}
		namedObj := named.Obj()
		gotName := namedObj.Name()
		if gotName != keyTypeName {
			pass.Reportf(call.Args[1].Pos(),
				"%s: context key must be the public type %s (type %s string), not %q",
				ruleKey, keyTypeName, keyTypeName, gotName)
			return true
		}
		if basic, ok := named.Underlying().(*types.Basic); !ok || basic.Kind() != types.String {
			pass.Reportf(call.Args[1].Pos(),
				"%s: %s must be a named string type", ruleKey, keyTypeName)
		}
		return true
	})
}

// checkKeyConsts — GID-167: const-значения ContextKey — string в snake_case,
// рядом с объявлением типа ContextKey (в одном файле).
func checkKeyConsts(pass *analysis.Pass, gd *ast.GenDecl, file, keyTypeFile *ast.File) {
	if gd.Tok != token.CONST {
		return
	}
	for _, spec := range gd.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range vs.Names {
			obj, ok := pass.TypesInfo.Defs[name].(*types.Const)
			if !ok {
				continue
			}
			named, ok := obj.Type().(*types.Named)
			if !ok {
				continue
			}
			namedObj := named.Obj()
			if namedObj.Name() != keyTypeName {
				continue
			}
			if keyTypeFile != nil && file != keyTypeFile {
				pass.Reportf(name.Pos(),
					"%s: %s values must be declared next to the %s type declaration (same file)",
					ruleKey, keyTypeName, keyTypeName)
			}
			objVal := obj.Val()
			if objVal.Kind() != constant.String {
				continue // не-string ловит checkKeyTypes по типу
			}
			if val := constant.StringVal(objVal); !snakeCase.MatchString(val) {
				pass.Reportf(name.Pos(),
					"%s: %s value must be a snake_case string, got %q", ruleKey, keyTypeName, val)
			}
		}
	}
}

func checkHelper(pass *analysis.Pass, file *ast.File, fn *ast.FuncDecl, structFile map[string]*ast.File) {
	name := fn.Name.Name
	if usesWithValue(pass, fn.Body) && !strings.HasPrefix(name, "ContextWith") {
		pass.Reportf(fn.Name.Pos(),
			"%s: function %q stores data in ctx. Fix: make it public and name it ContextWith<Name>",
			ruleNaming, name)
	}
	if usesCtxValue(pass, fn.Body) &&
		(!fn.Name.IsExported() || !strings.HasSuffix(name, "FromContext")) {
		pass.Reportf(fn.Name.Pos(),
			"%s: function %q reads data from ctx. Fix: make it public and name it <Name>FromContext",
			ruleNaming, name)
	}
	checkHelperFile(pass, file, fn, structFile)
}

// checkHelperFile: helper живёт в одном файле с сущностью <Name>.
func checkHelperFile(pass *analysis.Pass, file *ast.File, fn *ast.FuncDecl, structFile map[string]*ast.File) {
	entity := ""
	if rest, ok := strings.CutPrefix(fn.Name.Name, "ContextWith"); ok {
		entity = rest
	} else if rest, ok := strings.CutSuffix(fn.Name.Name, "FromContext"); ok {
		entity = rest
	}
	if entity == "" {
		return
	}
	declFile, declared := structFile[entity]
	if declared && declFile != file {
		pass.Reportf(fn.Name.Pos(),
			"%s: helper %q must live in the same file as the %q entity it stores into / reads from ctx",
			ruleNaming, fn.Name.Name, entity)
	}
}

func usesWithValue(pass *analysis.Pass, body *ast.BlockStmt) bool {
	return containsCall(body, func(call *ast.CallExpr) bool {
		return isWithValue(pass, call)
	})
}

func usesCtxValue(pass *analysis.Pass, body *ast.BlockStmt) bool {
	return containsCall(body, func(call *ast.CallExpr) bool {
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Value" {
			return false
		}
		return isContextType(pass.TypesInfo.TypeOf(sel.X))
	})
}

func containsCall(body *ast.BlockStmt, match func(*ast.CallExpr) bool) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		if call, ok := n.(*ast.CallExpr); ok && match(call) {
			found = true
			return false
		}
		return true
	})
	return found
}

func isWithValue(pass *analysis.Pass, call *ast.CallExpr) bool {
	f, ok := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
	if !ok {
		return false
	}
	pkg := f.Pkg()
	return pkg != nil && pkg.Path() == "context" && f.Name() == "WithValue"
}

func isContextType(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	return pkg != nil && pkg.Path() == "context" && obj.Name() == "Context"
}

func structFiles(pass *analysis.Pass) map[string]*ast.File {
	out := map[string]*ast.File{}
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, ok := ts.Type.(*ast.StructType); ok {
					out[ts.Name.Name] = file
				}
			}
		}
	}
	return out
}
