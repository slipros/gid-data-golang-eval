// Package nogetprefix implements rule GID-101: methods that fetch values do
// not carry the Get prefix (go-styleguide, "Method naming").
//
// The exception is generated code (protobuf and the like), where the Get
// prefix is part of the contract.
package nogetprefix

import (
	"go/ast"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-101"

// Analyzer — rule GID-101: the Get prefix is forbidden in method names. Fix: name getters without it (GetUser -> User).
var Analyzer = &analysis.Analyzer{
	Name: "gidnogetprefix",
	Doc:  ruleID + ": the Get prefix is forbidden in method names. Fix: name getters without it (GetUser -> User)",
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
			if hasGetPrefix(fn.Name.Name) {
				pass.Reportf(fn.Name.Pos(),
					"%s: method %q uses the Get prefix. Fix: name getters without it: %q",
					ruleID, fn.Name.Name, strings.TrimPrefix(fn.Name.Name, "Get"))
			}
		}
	}
	return nil, nil
}

// hasGetPrefix reports whether the name starts with the word Get: a bare "Get"
// or "Get" + a capitalized word ("GetJob"). Names like "Getaway", where get is
// part of another word, are not considered a violation.
func hasGetPrefix(name string) bool {
	if name == "Get" {
		return true
	}
	rest, ok := strings.CutPrefix(name, "Get")
	if !ok {
		return false
	}
	r, _ := utf8.DecodeRuneInString(rest)
	return unicode.IsUpper(r) || unicode.IsDigit(r)
}
