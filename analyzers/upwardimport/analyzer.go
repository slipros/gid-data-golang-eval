// Package upwardimport implements rule GID-131 (no-upward-import):
// a child package does not import its parent.
//
// The dependency direction: shared code is moved down, the parent imports
// the children, not the other way around. If the import path is a strict
// segment-wise prefix of the current package's path (pkgPath starts with
// impPath + "/"), the child package pulls in the parent — a dependency inversion.
//
// Self-import is impossible in Go; sibling packages and external modules do
// not match by definition (their path is not a prefix of the current
// package's path). The prefix is computed by path segments, not by string:
// "a/parentx" is not a child of "a/parent".
//
// Generated code (ast.IsGenerated) is skipped. LoadMode — TypesInfo
// (pass.Pkg.Path() is needed).
package upwardimport

import (
	"go/ast"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-131"

// Analyzer — rule GID-131: a child package does not import its parent.
var Analyzer = &analysis.Analyzer{
	Name: "gidupwardimport",
	Doc: ruleID + ": a child package must not import its parent. " +
		"Fix: invert the dependency, move shared code down and let the parent import children",
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, imp := range file.Imports {
			impPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			// A strict segment-wise prefix: pkgPath starts with impPath + "/".
			// This means impPath is the parent of the current package.
			if strings.HasPrefix(pkgPath, impPath+"/") {
				pass.Reportf(imp.Pos(),
					"%s: a child package imports its parent %s. "+
						"Fix: invert the dependency, move shared code down and let the parent import children",
					ruleID, impPath)
			}
		}
	}
	return nil, nil
}
