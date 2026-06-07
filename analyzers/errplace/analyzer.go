// Package errplace реализует правила размещения ошибок по слоям:
//
//   - GID-144 (giddomainerrors): все domain-ошибки живут в /domain/model.
//     service и usecase не объявляют и не создают ошибки — они обменивают
//     полученные ошибки на ошибки из model.
//   - GID-145 (giddalerrors): все dal-ошибки живут в /dal/entity.
//     Репозиторий обменивает ошибки подключения на ошибки из entity
//     (если обмена не произошло — пробрасывает исходную, это не создание).
//   - GID-169 (giderrfile): ошибки слоя живут в выделенном файле.
//     В корневых пакетах /domain/model и /dal/entity package-level
//     переменные типа error объявляются только в файле из settings.files
//     (default: error.go, errors.go). Уточняет GID-144/GID-145: те задают
//     слой-«дом» для ошибок, а GID-169 — конкретный файл внутри него.
//
// Запрещены вне разрешённого пакета: объявление package-level переменных
// типа error и вызовы конструкторов ошибок (errors.New, fmt.Errorf,
// errors.Errorf). Обмен и обогащение — errors.Wrap/WithStack/WithMessage
// (github.com/pkg/errors) и типизированные ошибки gderror — разрешены.
package errplace

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

// forbiddenCtors — конструкторы ошибок: пакет -> имена функций.
var forbiddenCtors = map[string]map[string]struct{}{
	"errors":                {"New": {}},
	"fmt":                   {"Errorf": {}},
	"github.com/pkg/errors": {"New": {}, "Errorf": {}},
}

// DomainAnalyzer — правило GID-144: all domain errors live in /domain/model. Fix: declare them in /domain/model.
var DomainAnalyzer = &analysis.Analyzer{
	Name: "giddomainerrors",
	Doc:  "GID-144: all domain errors live in /domain/model. Fix: declare them in /domain/model",
	Run: newRun(&config{
		ruleID:  "GID-144",
		tree:    []string{"domain"},
		allowed: []string{"domain", "model"},
		home:    "/domain/model",
	}),
}

// DALAnalyzer — правило GID-145: all dal errors live in /dal/entity. Fix: declare them in /dal/entity.
var DALAnalyzer = &analysis.Analyzer{
	Name: "giddalerrors",
	Doc:  "GID-145: all dal errors live in /dal/entity. Fix: declare them in /dal/entity",
	Run: newRun(&config{
		ruleID:  "GID-145",
		tree:    []string{"dal"},
		allowed: []string{"dal", "entity"},
		home:    "/dal/entity",
	}),
}

type config struct {
	ruleID  string
	tree    []string // дерево слоя, в котором правило действует
	allowed []string // подпуть, где ошибки разрешены
	home    string   // человекочитаемое место ошибок для сообщения
}

func newRun(cfg *config) func(*analysis.Pass) (any, error) {
	return func(pass *analysis.Pass) (any, error) {
		pkgPath := pass.Pkg.Path()
		if !pathseg.Contains(pkgPath, cfg.tree...) || pathseg.Contains(pkgPath, cfg.allowed...) {
			return nil, nil
		}
		for _, file := range pass.Files {
			if ast.IsGenerated(file) {
				continue
			}
			checkErrorVars(pass, cfg, file)
			checkErrorCtors(pass, cfg, file)
		}
		return nil, nil
	}
}

// checkErrorVars ищет package-level переменные, реализующие error.
func checkErrorVars(pass *analysis.Pass, cfg *config, file *ast.File) {
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if name.Name == "_" {
					continue
				}
				obj := pass.TypesInfo.Defs[name]
				if obj == nil || !implementsError(obj.Type()) {
					continue
				}
				pass.Reportf(name.Pos(),
					"%s: error %q is declared in %q. Fix: keep this layer's errors in %s",
					cfg.ruleID, name.Name, pass.Pkg.Path(), cfg.home)
			}
		}
	}
}

// checkErrorCtors ищет вызовы конструкторов ошибок.
func checkErrorCtors(pass *analysis.Pass, cfg *config, file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		fn := typeutil.Callee(pass.TypesInfo, call)
		f, ok := fn.(*types.Func)
		if !ok || f.Pkg() == nil {
			return true
		}
		fPkg := f.Pkg()
		names, ok := forbiddenCtors[fPkg.Path()]
		if !ok {
			return true
		}
		if _, ok := names[f.Name()]; !ok {
			return true
		}
		pass.Reportf(call.Pos(),
			"%s: creating an error via %s.%s is forbidden. Fix: exchange it for an error from %s (Wrap/WithStack are allowed)",
			cfg.ruleID, fPkg.Name(), f.Name(), cfg.home)
		return true
	})
}

func implementsError(t types.Type) bool {
	errObj := types.Universe.Lookup("error")
	errType := errObj.Type()
	errIface, ok := errType.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	return types.Implements(t, errIface)
}
