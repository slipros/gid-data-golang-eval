// Package embedmutex реализует правило GID-178 (gidembedmutex):
// запрет встраивания (анонимное поле) sync.Mutex / sync.RWMutex
// (а также указателей на них) в структуры.
//
// Встраивание мьютекса промотирует его методы Lock/Unlock в публичный
// API типа: внешний код может залочить чужой мьютекс. Мьютекс хранится
// именованным неэкспортируемым полем (mu sync.Mutex), оставаясь деталью
// реализации.
//
// Детект — через go/types (pass.TypesInfo), а не по тексту селектора,
// чтобы устойчиво работать при алиасах импорта пакета sync. Анонимное
// поле структуры, тип которого после снятия указателя — именованный тип
// Mutex или RWMutex из стандартного пакета "sync". Именованное поле любого
// вида допустимо. Сгенерированный код пропускается.
package embedmutex

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-178"

// Analyzer — правило GID-178: не встраивайте sync.Mutex/sync.RWMutex — храните именованным полем (mu sync.Mutex).
var Analyzer = &analysis.Analyzer{
	Name: "gidembedmutex",
	Doc:  ruleID + ": не встраивайте sync.Mutex/sync.RWMutex — храните именованным полем (mu sync.Mutex)",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			st, ok := n.(*ast.StructType)
			if !ok || st.Fields == nil {
				return true
			}
			for _, field := range st.Fields.List {
				// Анонимное (встроенное) поле — без имён.
				if len(field.Names) != 0 {
					continue
				}
				name, ok := embeddedMutexName(pass.TypesInfo.TypeOf(field.Type))
				if !ok {
					continue
				}
				pass.Reportf(field.Pos(),
					"%s: sync.%s встроен в структуру — храните мьютекс именованным полем (mu sync.Mutex), "+
						"иначе Lock/Unlock попадают в API типа",
					ruleID, name)
			}
			return true
		})
	}
	return nil, nil
}

// embeddedMutexName возвращает имя типа мьютекса ("Mutex" или "RWMutex"),
// если t (после снятия указателя) — именованный тип Mutex/RWMutex из пакета
// "sync". Иначе ok == false.
func embeddedMutexName(t types.Type) (string, bool) {
	if t == nil {
		return "", false
	}
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return "", false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	if pkg == nil || pkg.Path() != "sync" {
		return "", false
	}
	switch obj.Name() {
	case "Mutex", "RWMutex":
		return obj.Name(), true
	default:
		return "", false
	}
}
