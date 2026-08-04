// Package srcfile answers questions about the source file a node came from.
//
// The one question so far is IsTest. A test lives in the same package as the
// code under test (GID-250 — the team convention is the opposite of the
// standard testpackage linter), so every rule written for production code sees
// _test.go files too, and judges scaffolding it was never meant to judge: a
// double copying the method names and signatures of the interface it fakes, a
// sentinel error that has no place in /domain/model, a harness holding several
// services at once, a must-helper that panics where t is not available yet.
// Such a rule calls IsTest and skips the file (incident 2026-08-04,
// advertising-api: 107 of 138 diagnostics of the naming and layer rules landed
// on test doubles).
//
// A rule that judges the test itself (GID-250 testpackage, subtest naming)
// does the opposite and must not use this.
package srcfile

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// IsTest reports whether file is a _test.go file.
func IsTest(pass *analysis.Pass, file *ast.File) bool {
	tokenFile := pass.Fset.File(file.Pos())

	return tokenFile != nil && strings.HasSuffix(tokenFile.Name(), "_test.go")
}

// IsTestBinaryPkg reports whether the package under analysis is the synthesized
// test binary ("pkg.test", holding the generated _testmain.go). A rule that
// reports per package — its placement in the tree, its layout — must skip it:
// the package has no source directory of its own, and its production twin is
// judged anyway.
func IsTestBinaryPkg(pass *analysis.Pass) bool {
	return strings.HasSuffix(pass.Pkg.Path(), ".test")
}
