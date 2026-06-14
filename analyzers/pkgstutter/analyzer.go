// Package pkgstutter implements rule GID-193 (no-pkg-stutter): an exported
// top-level symbol (type, function, var, const) must not start with the
// package name. From outside such a symbol reads with a stutter:
// widget.WidgetOptions, widget.WidgetService — the package name already gives
// context, the prefix is redundant. widget.Options, widget.Service suffice.
//
// Comparison is done at the CamelCase word boundary: the package name must
// match the symbol's first word entirely. Package widget matches
// WidgetOptions/WidgetCount, but package log does NOT match Logger (Logger
// starts with the word "Logger", not "Log"), and package conv does NOT match
// Convert (the word is "Convert", not "Conv").
//
// Exceptions (not matched):
//   - New* constructors — our GID-104 requires New<Entity>, the conflict is resolved in its favor;
//   - methods (with a receiver) and unexported symbols;
//   - package main.
//
// Generated files (ast.IsGenerated) are skipped.
// LoadMode — Syntax: no types needed, the package name and the AST suffice.
package pkgstutter

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-193"

// Analyzer — rule GID-193: an exported symbol does not repeat the package name (widget.WidgetOptions).
var Analyzer = &analysis.Analyzer{
	Name: "gidpkgstutter",
	Doc:  ruleID + ": an exported symbol must not repeat the package name; from outside it stutters (widget.WidgetOptions). Fix: drop the prefix",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	pkgName := pass.Pkg.Name()
	if pkgName == "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				checkFunc(pass, pkgName, d)
			case *ast.GenDecl:
				checkGenDecl(pass, pkgName, d)
			}
		}
	}
	return nil, nil
}

func checkFunc(pass *analysis.Pass, pkgName string, fn *ast.FuncDecl) {
	if fn.Recv != nil {
		return // a method — it has a receiver, the name reads as value.Method
	}
	name := fn.Name.Name
	if strings.HasPrefix(name, "New") {
		return // a New* constructor — GID-104 takes precedence
	}
	report(pass, pkgName, name, fn.Name.Pos())
}

func checkGenDecl(pass *analysis.Pass, pkgName string, gd *ast.GenDecl) {
	for _, spec := range gd.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			report(pass, pkgName, s.Name.Name, s.Name.Pos())
		case *ast.ValueSpec:
			for _, ident := range s.Names {
				report(pass, pkgName, ident.Name, ident.Pos())
			}
		}
	}
}

// report emits a diagnostic if name is exported and its first CamelCase word
// matches the package name (case-insensitively).
func report(pass *analysis.Pass, pkgName, name string, pos token.Pos) {
	if !ast.IsExported(name) {
		return
	}
	if !stutters(pkgName, name) {
		return
	}
	suffix := name[len(pkgName):]
	pass.Reportf(pos,
		"%s: %s repeats the package name %s. Fix: from outside it is %s.%s; drop the prefix",
		ruleID, name, pkgName, pkgName, suffix)
}

// stutters reports whether the symbol starts with the package name as a
// separate CamelCase word. The comparison is case-insensitive, but the word
// boundary is respected: after the prefix of length len(pkgName) a new capital
// letter (the next word) must begin, otherwise the package name is just part
// of another word (log → Logger).
func stutters(pkgName, name string) bool {
	if len(name) <= len(pkgName) {
		return false // an exact match or shorter — there is no next word
	}
	if !strings.EqualFold(name[:len(pkgName)], pkgName) {
		return false
	}
	// The next rune must be the start of a new CamelCase word — a capital
	// letter. If lowercase, the package name is only a prefix of a word.
	next := rune(name[len(pkgName)])
	return unicode.IsUpper(next)
}
