// Package entitygroup implements rule GID-157: an entity's code is a single
// block. All of an entity's functions live in the file of its declaration, in
// the order type -> constructor -> methods. Functions of different entities are
// not mixed into one pile, and no foreign declaration (a free function, an
// unrelated type) splits the block: free helpers go either above the first type
// or below the entity's last method — never in between.
//
// A constructor is recognised by what it builds, not by a name template: a
// receiverless function is the entity's code when it is named New<Entity>, or
// when its first result is T or *T for a struct T of the package. So an
// unexported factory (newPoolStatsCollector), a second constructor of one
// entity (NewLoggerByEntry) and one returning an interface the entity
// implements (NewWithARGS() Logger) belong to the block instead of splitting it.
//
// A _test.go file is not judged at all: its composition — a table, its
// fixtures, a builder returning the entity under test — is the test's own.
package entitygroup

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/modlayout"
)

const ruleID = "GID-157"

const (
	kindType declKind = iota
	kindCtor
	kindMethod
	kindForeign
)

// Analyzer — rule GID-157: an entity's code is a single block: type, constructor, methods.
var Analyzer = &analysis.Analyzer{
	Name: "gidentitygroup",
	Doc:  ruleID + ": an entity's code must be one block (type, constructor, methods) without interleaving entities or free functions. Fix: keep the entity's declarations together",
	Run:  run,
}

// ownedDecl — a declaration belonging to an entity; kindForeign marks a
// declaration owned by no entity (a free function, a non-struct type).
type ownedDecl struct {
	entity string
	kind   declKind
	name   *ast.Ident
}

type declKind int

func run(pass *analysis.Pass) (any, error) {
	typeFile := structFiles(pass)
	// In a library the cross-file half of the rule does not apply: a client with
	// one big type spreads its methods over topic files (domain.go, lineage.go)
	// the way go-github does, and that is the idiom, not a mess. What still
	// holds there is the contiguity of a block inside its own file.
	crossFile := modlayout.IsServiceModule(pass)
	for _, file := range pass.Files {
		if ast.IsGenerated(file) || isTestFile(pass, file) {
			continue
		}
		checkFile(pass, file, typeFile, crossFile)
	}
	return nil, nil
}

// isTestFile reports whether the file is a _test.go file. Tests live in the
// same package (GID-250), but their composition is their own: a table, its
// fixtures and a builder returning the entity under test belong to the test
// file, not to the file where the entity is declared.
func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	tokenFile := pass.Fset.File(file.Pos())

	return tokenFile != nil && strings.HasSuffix(tokenFile.Name(), "_test.go")
}

func checkFile(pass *analysis.Pass, file *ast.File, typeFile map[string]*ast.File, crossFile bool) {
	owned := ownedDecls(file, typeFile)

	typeIdx := map[string]int{}
	methodIdx := map[string]int{}
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for i, d := range owned {
		switch d.kind {
		case kindType:
			typeIdx[d.entity] = i
		case kindMethod:
			// The first method closes the constructor section: an entity may have
			// several constructors (NewX, newDefaultX, NewXByY) and they all sit
			// together under the type declaration, above the methods.
			if _, seen := methodIdx[d.entity]; !seen {
				methodIdx[d.entity] = i
			}
		}
	}

	// Methods and the constructor live in the entity's declaration file.
	if crossFile {
		reportForeignFile(pass, file, owned, typeFile)
	}

	// The order inside a file: type -> constructor -> methods.
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for i, d := range owned {
		ti, hasType := typeIdx[d.entity]
		switch d.kind {
		case kindCtor:
			if hasType && i < ti {
				pass.Reportf(d.name.Pos(),
					"%s: constructor %q must be placed right below the %q type declaration", ruleID, d.name.Name, d.entity)

				continue
			}
			if mi, hasMethod := methodIdx[d.entity]; hasMethod && i > mi {
				pass.Reportf(d.name.Pos(),
					"%s: constructor %q sits below the methods of %q. Fix: keep every constructor of an entity together "+
						"under its type declaration, above the methods", ruleID, d.name.Name, d.entity)
			}
		case kindMethod:
			if hasType && i < ti {
				pass.Reportf(d.name.Pos(),
					"%s: method %q must be placed below the %q type declaration", ruleID, d.name.Name, d.entity)
			}
		}
	}

	// Interleaving: the entity block must be contiguous.
	seen := map[string]struct{}{}
	last := ""
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for _, d := range owned {
		if d.kind == kindForeign || d.entity == last {
			continue
		}
		if _, ok := seen[d.entity]; ok {
			pass.Reportf(d.name.Pos(),
				"%s: entity %q code is interleaved with other entities. Fix: keep the entity block contiguous",
				ruleID, d.entity)
		}
		seen[last] = struct{}{}
		last = d.entity
	}

	reportSplits(pass, owned)
}

// reportForeignFile — a method or a constructor declared away from the file
// holding its type. Service-only: a library client spreads the methods of one
// type over topic files by design.
func reportForeignFile(pass *analysis.Pass, file *ast.File, owned []ownedDecl, typeFile map[string]*ast.File) {
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for _, d := range owned {
		if d.kind == kindType || d.kind == kindForeign {
			continue
		}
		declFile, ok := typeFile[d.entity]
		if ok && declFile != file {
			pass.Reportf(d.name.Pos(),
				"%s: %q belongs to entity %q. Fix: keep the entity's code in the file where it is declared",
				ruleID, d.name.Name, d.entity)
		}
	}
}

// reportSplits — a free function or an unrelated type sitting between an
// entity's own declarations tears its block apart. Const and var blocks are
// left to GID-130, which already fixes their place in the file.
func reportSplits(pass *analysis.Pass, owned []ownedDecl) {
	first, lastIdx := entityBounds(owned)

	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for i, d := range owned {
		if d.kind != kindForeign {
			continue
		}
		// The innermost block wins: several entities may span the same index,
		// and the nearest one is the block the reader sees torn apart.
		split, from := "", -1
		for entity, start := range first {
			if start < i && i < lastIdx[entity] && start > from {
				split, from = entity, start
			}
		}
		if split == "" {
			continue
		}
		pass.Reportf(d.name.Pos(),
			"%s: %q splits the %q entity block. Fix: move it above the first type or below the entity's last method",
			ruleID, d.name.Name, split)
	}
}

// entityBounds — the index of the first and of the last declaration of every
// entity present in the file.
func entityBounds(owned []ownedDecl) (first, last map[string]int) {
	first, last = map[string]int{}, map[string]int{}

	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for i, d := range owned {
		if d.kind == kindForeign {
			continue
		}
		if _, ok := first[d.entity]; !ok {
			first[d.entity] = i
		}
		last[d.entity] = i
	}

	return first, last
}

// ownedDecls — the sequence of the file's declarations with their entities.
func ownedDecls(file *ast.File, typeFile map[string]*ast.File) []ownedDecl {
	var out []ownedDecl
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, ok := ts.Type.(*ast.StructType); ok {
					out = append(out, ownedDecl{entity: ts.Name.Name, kind: kindType, name: ts.Name})
					continue
				}
				out = append(out, ownedDecl{kind: kindForeign, name: ts.Name})
			}
		case *ast.FuncDecl:
			if d.Recv != nil {
				if recv := recvTypeName(d); recv != "" {
					out = append(out, ownedDecl{entity: recv, kind: kindMethod, name: d.Name})
				}
				continue
			}
			if entity, ok := ctorEntity(d, typeFile); ok {
				out = append(out, ownedDecl{entity: entity, kind: kindCtor, name: d.Name})
				continue
			}
			out = append(out, ownedDecl{kind: kindForeign, name: d.Name})
		}
	}
	return out
}

// ctorEntity — the entity a receiverless function constructs. Only a New*/new*
// function qualifies — that is how a constructor is named in Go, and the prefix
// keeps an ordinary helper returning the type (pairOf() *namingPair) out: such a
// helper sits wherever it is convenient and must not stretch the entity's block
// across the whole file.
//
// The entity is taken from the name (New<Entity>, which also covers a
// constructor returning an interface the entity implements — NewWithARGS()
// Logger) or, failing that, from the first result: T or *T for a struct of the
// package. So newPoolStatsCollector and NewLoggerByEntry are the entity's code.
func ctorEntity(fn *ast.FuncDecl, typeFile map[string]*ast.File) (string, bool) {
	name, isCtor := ctorName(fn.Name.Name)
	if !isCtor {
		return "", false
	}

	if _, declared := typeFile[name]; declared {
		return name, true
	}

	if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		return "", false
	}
	t := fn.Type.Results.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	ident, ok := t.(*ast.Ident)
	if !ok {
		return "", false
	}
	if _, declared := typeFile[ident.Name]; !declared {
		return "", false
	}
	return ident.Name, true
}

// ctorName — the part of a constructor's name after the New/new prefix. A bare
// New is a constructor too — the idiom of a library package (batch.New), where
// GID-104 does not ask for the entity in the name — and its entity is then read
// from the result type.
func ctorName(fnName string) (string, bool) {
	for _, prefix := range [...]string{"New", "new"} {
		if rest, ok := strings.CutPrefix(fnName, prefix); ok {
			return rest, true
		}
	}

	return "", false
}

// structFiles — the declaration file of each struct in the package.
func structFiles(pass *analysis.Pass) map[string]*ast.File {
	out := map[string]*ast.File{}
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, ok := ts.Type.(*ast.StructType); ok {
					out[ts.Name.Name] = file
				}
			}
		}
	}
	return out
}

func recvTypeName(fn *ast.FuncDecl) string {
	t := fn.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	if ident, ok := t.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}
