// Package nobatch implements rule GID-102: the word Batch is not used in
// method names (go-styleguide, "Method naming"). A method working with
// multiple entities is named like the single-entity one, but in the plural:
// CreateJob -> CreateJobs, not CreateBatchJobs / BatchCreate.
package nobatch

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-102"

// Analyzer — rule GID-102: the word Batch is forbidden in method names. Fix: use a plural instead (CreateJob -> CreateJobs).
var Analyzer = &analysis.Analyzer{
	Name: "gidnobatch",
	Doc:  ruleID + ": the word Batch is forbidden in method names. Fix: use a plural instead (CreateJob -> CreateJobs)",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil {
				continue
			}
			if strings.Contains(fn.Name.Name, "Batch") {
				pass.Reportf(fn.Name.Pos(),
					"%s: method %q contains the word Batch. Fix: use a plural instead (CreateJob -> CreateJobs)",
					ruleID, fn.Name.Name)
			}
		}
	}
	return nil, nil
}
