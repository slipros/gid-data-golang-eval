package enumstring

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleIDBased = "GID-123"

// BasedAnalyzer implements GID-123: an enum is a named string-based type,
// not a bare string/int. Applies in /domain/model/**, /dal/entity/** and
// /event/dto/**.
var BasedAnalyzer = &analysis.Analyzer{
	Name: "gidenumbased",
	Doc:  ruleIDBased + ": an enum must be a named type based on string, not a bare string/int. Fix: declare a named string type",
	Run:  runBased,
}

func runBased(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	inScope := pathseg.Contains(pkgPath, "domain", "model") ||
		pathseg.Contains(pkgPath, "dal", "entity") ||
		pathseg.Contains(pkgPath, "event", "dto")
	if !inScope {
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

// checkTypeDecl catches an alias to a basic type and an int-enum (≥2 const values).
func checkTypeDecl(pass *analysis.Pass, gd *ast.GenDecl, intEnums map[*types.Named]struct{}) {
	for _, spec := range gd.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		// Check 1: alias to a basic string/int.
		if ts.Assign != token.NoPos && isBasicStringOrInt(pass.TypesInfo.TypeOf(ts.Type)) {
			pass.Reportf(ts.Name.Pos(),
				"%s: enum %s must be a named type, not an alias (type %s = ...). Fix: declare type %s string "+
					"and retype the constants",
				ruleIDBased, ts.Name.Name, ts.Name.Name, ts.Name.Name)
			continue
		}
		// Check 2: a named int type with ≥2 const values.
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
				"%s: enum %s must be based on string, not int. Fix: declare type %s string and give the "+
					"constants string values",
				ruleIDBased, ts.Name.Name, ts.Name.Name)
		}
	}
}

// checkUntypedStringConstGroup catches a group of ≥2 untyped string consts
// in one const block (a single GenDecl). Reports once per group.
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
			// Untyped string == universe string (not a named
			// type). The default type of an untyped const is Basic String;
			// an explicit untyped one is UntypedString.
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
			"%s: a group of string constants. Fix: declare a named string type (enum)", ruleIDBased)
	}
}

// intEnumTypes — the package's named types with an underlying integer and ≥2 const values.
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

// isBasicStringOrInt reports whether the type is a universe string or integer.
func isBasicStringOrInt(t types.Type) bool {
	basic, ok := t.(*types.Basic)
	if !ok {
		return false
	}
	return basic.Kind() == types.String || basic.Info()&types.IsInteger != 0
}
