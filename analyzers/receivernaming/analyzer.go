// Package receivernaming реализует правило GID-103: ресивер именуется
// первой буквой названия структуры в нижнем регистре; для слайс-типов —
// двумя буквами (type Snapshots []Snapshot -> ss).
//
// Исключения стайлгайда: в validate-пакетах ресивер v, в handler-пакетах — h.
package receivernaming

import (
	"go/ast"
	"go/types"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-103"

// Analyzer — правило GID-103: ресивер — первая буква типа (две — для слайс-типов).
var Analyzer = &analysis.Analyzer{
	Name: "gidreceiver",
	Doc:  ruleID + ": a receiver is the lowercase first letter of the type, two for slice types. Fix: rename the receiver",
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
		return // безымянный ресивер не именуется
	}
	got := recv.Names[0].Name
	typeName, isSlice := recvType(pass, recv)
	if typeName == "" {
		return
	}
	want := expected(typeName, isSlice)
	if got == want || allowedException(pass.Pkg.Path(), got) {
		return
	}
	pass.Reportf(recv.Names[0].Pos(),
		"%s: receiver of type %s is named %q. Fix: use the lowercase first letter of the type (two for slice types), got %q",
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
		return letter + letter
	}
	return letter
}

// allowedException: v в validate-пакетах, h в handler-пакетах.
func allowedException(pkgPath, got string) bool {
	switch got {
	case "v":
		return pathseg.Contains(pkgPath, "validate") || strings.HasSuffix(pkgPath, "validator")
	case "h":
		return pathseg.Contains(pkgPath, "handler")
	}
	return false
}
