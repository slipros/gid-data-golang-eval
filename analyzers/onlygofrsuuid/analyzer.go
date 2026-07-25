// Package onlygofrsuuid implements rule GID-137: only the
// github.com/gofrs/uuid library is used for UUIDs (go-styleguide,
// "Identifiers"). Importing alternative uuid libraries is forbidden.
package onlygofrsuuid

import (
	"go/ast"
	"strconv"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const (
	ruleID     = "GID-137"
	allowedPkg = "github.com/gofrs/uuid"
)

// deniedPkgs — known alternative uuid libraries. Each is matched at any major
// version (github.com/google/uuid/v2 counts) through pathseg.SameLibrary.
var deniedPkgs = []string{
	"github.com/google/uuid",
	"github.com/satori/go.uuid",
	"github.com/pborman/uuid",
	"github.com/hashicorp/go-uuid",
	"github.com/twinj/uuid",
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
			if denied, ok := deniedLibrary(path); ok {
				pass.Reportf(imp.Pos(),
					"%s: importing %q is forbidden (%s). Fix: use %s for UUID",
					ruleID, path, denied, allowedPkg)
			}
		}
	}
	return nil, nil
}

// deniedLibrary reports whether path is one of the alternative uuid libraries
// (at any major version), returning the library it matched.
func deniedLibrary(path string) (string, bool) {
	for _, denied := range deniedPkgs {
		if pathseg.SameLibrary(path, denied) {
			return denied, true
		}
	}
	return "", false
}
