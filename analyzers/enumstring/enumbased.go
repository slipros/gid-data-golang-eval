package enumstring

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleIDBased = "GID-123"

// BasedAnalyzer реализует GID-123: enum — именованный тип на основе string,
// не голый string/int. Действует только в /domain/model/** и /dal/entity/**.
var BasedAnalyzer = &analysis.Analyzer{
	Name: "gidenumbased",
	Doc:  ruleIDBased + ": enum — именованный тип на основе string, не голый string/int",
	Run:  runBased,
}

func runBased(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	if !pathseg.Contains(pkgPath, "domain", "model") && !pathseg.Contains(pkgPath, "dal", "entity") {
		return nil, nil
	}

	intEnums := intEnumTypes(pass)
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			switch gd.Tok {
			case token.TYPE:
				checkTypeDecl(pass, gd, intEnums)
			case token.CONST:
				checkUntypedStringConstGroup(pass, gd)
			}
		}
	}
	return nil, nil
}

// checkTypeDecl ловит alias на basic-тип и int-enum (≥2 const-значений).
func checkTypeDecl(pass *analysis.Pass, gd *ast.GenDecl, intEnums map[*types.Named]struct{}) {
	for _, spec := range gd.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		// Проверка 1: alias на basic string/int.
		if ts.Assign != token.NoPos && isBasicStringOrInt(pass.TypesInfo.TypeOf(ts.Type)) {
			pass.Reportf(ts.Name.Pos(),
				"%s: enum %s — именованный тип, не alias (type %s = ...)",
				ruleIDBased, ts.Name.Name, ts.Name.Name)
			continue
		}
		// Проверка 2: именованный int-тип с ≥2 const-значениями.
		obj, ok := pass.TypesInfo.Defs[ts.Name].(*types.TypeName)
		if !ok {
			continue
		}
		named, ok := obj.Type().(*types.Named)
		if !ok {
			continue
		}
		if _, isIntEnum := intEnums[named]; isIntEnum {
			pass.Reportf(ts.Name.Pos(),
				"%s: enum %s строится на string, не int", ruleIDBased, ts.Name.Name)
		}
	}
}

// checkUntypedStringConstGroup ловит группу из ≥2 нетипизированных string-const
// в одном const-блоке (один GenDecl). Репорт один раз на группу.
func checkUntypedStringConstGroup(pass *analysis.Pass, gd *ast.GenDecl) {
	var firstPos token.Pos
	count := 0
	for _, spec := range gd.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range vs.Names {
			if name.Name == "_" {
				continue
			}
			c, ok := pass.TypesInfo.Defs[name].(*types.Const)
			if !ok {
				continue
			}
			// Нетипизированный string == universe string (не именованный
			// тип). Дефолтный тип нетипизированной const — Basic String;
			// явная нетипизированная — UntypedString.
			basic, ok := c.Type().(*types.Basic)
			if !ok {
				continue
			}
			if basic.Kind() != types.String && basic.Kind() != types.UntypedString {
				continue
			}
			count++
			if firstPos == token.NoPos {
				firstPos = name.Pos()
			}
		}
	}
	if count >= 2 {
		pass.Reportf(firstPos,
			"%s: группа string-констант — заведите именованный string-тип (enum)", ruleIDBased)
	}
}

// intEnumTypes — именованные типы пакета с underlying integer и ≥2 const-значений.
func intEnumTypes(pass *analysis.Pass) map[*types.Named]struct{} {
	counts := map[*types.Named]int{}
	for _, obj := range pass.TypesInfo.Defs {
		c, ok := obj.(*types.Const)
		if !ok {
			continue
		}
		named, ok := c.Type().(*types.Named)
		if !ok {
			continue
		}
		namedObj := named.Obj()
		if namedObj.Pkg() != pass.Pkg {
			continue
		}
		basic, ok := named.Underlying().(*types.Basic)
		if !ok || basic.Info()&types.IsInteger == 0 {
			continue
		}
		counts[named]++
	}
	out := map[*types.Named]struct{}{}
	for named, n := range counts {
		if n >= 2 {
			out[named] = struct{}{}
		}
	}
	return out
}

// isBasicStringOrInt сообщает, является ли тип universe string или integer.
func isBasicStringOrInt(t types.Type) bool {
	basic, ok := t.(*types.Basic)
	if !ok {
		return false
	}
	return basic.Kind() == types.String || basic.Info()&types.IsInteger != 0
}
