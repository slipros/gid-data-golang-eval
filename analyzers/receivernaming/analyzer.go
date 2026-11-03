// Package receivernaming implements rule GID-103: a receiver is named by the
// lowercase first letter of the type name; for slice types — by the initials of
// the first two words (ConsentProfiles -> cp), or a doubled initial (Snapshots -> ss).
//
// No exceptions: the rule applies uniformly, including validate and handler
// packages.
package receivernaming

import (
	"go/ast"
	"go/types"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-103"

// Analyzer checks receiver names against type initials.
var Analyzer = &analysis.Analyzer{
	Name: "gidreceiver",
	Doc:  ruleID + ": a receiver is the lowercase first letter of the type; slices use the first two word initials (a doubled initial for one word). Fix: rename the receiver",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}
			checkReceiver(pass, fn)
		}
	}
	return nil, nil
}

func checkReceiver(pass *analysis.Pass, fn *ast.FuncDecl) {
	recv := fn.Recv.List[0]
	if len(recv.Names) == 0 || recv.Names[0].Name == "_" {
		return // an unnamed receiver has no name to check
	}
	got := recv.Names[0].Name
	typeName, isSlice := recvType(pass, recv)
	if typeName == "" {
		return
	}
	want := expected(typeName, isSlice)
	if got == want {
		return
	}
	pass.Reportf(recv.Names[0].Pos(),
		"%s: receiver of type %s is named %q. Fix: use the lowercase first letter of the type (first two word initials for slices; double the initial for one word), got %q",
		ruleID, typeName, want, got)
}

func recvType(pass *analysis.Pass, recv *ast.Field) (string, bool) {
	t := pass.TypesInfo.TypeOf(recv.Type)
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return "", false
	}
	_, isSlice := named.Underlying().(*types.Slice)
	obj := named.Obj()
	return obj.Name(), isSlice
}

func expected(typeName string, isSlice bool) string {
	r, _ := utf8.DecodeRuneInString(typeName)
	letter := string(unicode.ToLower(r))
	if isSlice {
		runes := []rune(typeName)
		for i := 1; i < len(runes); i++ {
			if !unicode.IsUpper(runes[i]) {
				continue
			}
			if unicode.IsUpper(runes[i-1]) && i+2 == len(runes) && runes[i+1] == 's' {
				continue // A plural acronym such as URLs is still one word.
			}
			if unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1]) ||
				(unicode.IsUpper(runes[i-1]) && i+1 < len(runes) && unicode.IsLower(runes[i+1])) {
				return letter + string(unicode.ToLower(runes[i]))
			}
		}
		return letter + letter
	}
	return letter
}
