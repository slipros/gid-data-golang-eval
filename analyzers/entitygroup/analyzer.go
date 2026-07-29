// Package entitygroup implements rule GID-157: an entity's code is a single
// block. All of an entity's functions live in the file of its declaration, in
// the order type -> the New<Entity> constructor -> methods. Functions of
// different entities are not mixed into one pile, and no foreign declaration
// (a free function, an unrelated type) splits the block: free helpers go either
// above the first type or below the entity's last method — never in between.
package entitygroup

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
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
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		checkFile(pass, file, typeFile)
	}
	return nil, nil
}

func checkFile(pass *analysis.Pass, file *ast.File, typeFile map[string]*ast.File) {
	owned := ownedDecls(file, typeFile)

	typeIdx := map[string]int{}
	ctorIdx := map[string]int{}
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for i, d := range owned {
		switch d.kind {
		case kindType:
			typeIdx[d.entity] = i
		case kindCtor:
			ctorIdx[d.entity] = i
		}
	}

	// Methods and the constructor live in the entity's declaration file.
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

	// The order inside a file: type -> constructor -> methods.
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for i, d := range owned {
		ti, hasType := typeIdx[d.entity]
		switch d.kind {
		case kindCtor:
			if hasType && i < ti {
				pass.Reportf(d.name.Pos(),
					"%s: constructor %q must be placed right below the %q type declaration", ruleID, d.name.Name, d.entity)
			}
		case kindMethod:
			if hasType && i < ti {
				pass.Reportf(d.name.Pos(),
					"%s: method %q must be placed below the %q type declaration", ruleID, d.name.Name, d.entity)
			}
			if ci, hasCtor := ctorIdx[d.entity]; hasCtor && i < ci {
				pass.Reportf(d.name.Pos(),
					"%s: method %q must be placed below the New%s constructor", ruleID, d.name.Name, d.entity)
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
			if entity, ok := strings.CutPrefix(d.Name.Name, "New"); ok && entity != "" {
				if _, declared := typeFile[entity]; declared {
					out = append(out, ownedDecl{entity: entity, kind: kindCtor, name: d.Name})
					continue
				}
			}
			out = append(out, ownedDecl{kind: kindForeign, name: d.Name})
		}
	}
	return out
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
