// Package onlygofrsuuid implements rule GID-137: only the
// github.com/gofrs/uuid library is used for UUIDs (go-styleguide,
// "Identifiers"). Importing alternative uuid libraries is forbidden.
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

// deniedPkgs — known alternative uuid libraries.
var deniedPkgs = map[string]struct{}{
	"github.com/google/uuid":       {},
	"github.com/satori/go.uuid":    {},
	"github.com/pborman/uuid":      {},
	"github.com/hashicorp/go-uuid": {},
	"github.com/twinj/uuid":        {},
}

// Analyzer — rule GID-137: only github.com/gofrs/uuid is allowed for UUIDs.
var Analyzer = &analysis.Analyzer{
	Name: "gidonlygofrsuuid",
	Doc:  ruleID + ": for UUID only this library is allowed: " + allowedPkg,
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
					"%s: importing %q is forbidden. Fix: use %s for UUID",
					ruleID, path, allowedPkg)
			}
		}
	}
	return nil, nil
}
