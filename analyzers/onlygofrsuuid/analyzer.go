// Package onlygofrsuuid реализует правило GID-137: для работы с UUID
// используется только библиотека github.com/gofrs/uuid (go-styleguide,
// «Идентификаторы»). Импорт альтернативных uuid-библиотек запрещён.
package onlygofrsuuid

import (
	"go/ast"
	"strconv"

	"golang.org/x/tools/go/analysis"
)

const (
	ruleID     = "GID-137"
	allowedPkg = "github.com/gofrs/uuid"
)

// deniedPkgs — известные альтернативные uuid-библиотеки.
var deniedPkgs = map[string]struct{}{
	"github.com/google/uuid":       {},
	"github.com/satori/go.uuid":    {},
	"github.com/pborman/uuid":      {},
	"github.com/hashicorp/go-uuid": {},
	"github.com/twinj/uuid":        {},
}

// Analyzer — правило GID-137: для UUID разрешена только github.com/gofrs/uuid.
var Analyzer = &analysis.Analyzer{
	Name: "gidonlygofrsuuid",
	Doc:  ruleID + ": для UUID разрешена только библиотека " + allowedPkg,
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, imp := range file.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			if _, denied := deniedPkgs[path]; denied {
				pass.Reportf(imp.Pos(),
					"%s: импорт %q запрещён — для UUID используйте %s",
					ruleID, path, allowedPkg)
			}
		}
	}
	return nil, nil
}
