// Package chandir реализует правило GID-189 (Google: channel direction):
// параметр функции/метода с типом двунаправленного канала chan T обязан
// указывать направление — <-chan T для чтения или chan<- T для записи.
// Двунаправленный канал в сигнатуре скрывает намерение вызывающего и
// допускает ошибочное использование (чтение из канала, который должен
// только заполняться, и наоборот).
//
// Что матчится:
//   - параметр функции func f(ch chan int);
//   - параметр метода func (s S) m(ch chan int);
//   - параметр функционального литерала func(ch chan int) { ... }.
//     Матчим именно литеральный chan T в позиции параметра.
//
// Что НЕ матчится:
//   - направленные каналы <-chan T / chan<- T — направление уже указано;
//   - возвращаемые значения (func f() chan T) — владельцу канала бывает
//     нужен двунаправленный, решение за review;
//   - поля структур (chan T) — направление задаётся при передаче в функцию;
//   - локальные переменные (var ch chan T) — это создание канала;
//   - параметр с именованным типом-каналом (type Pipe chan int; func f(p Pipe)) —
//     именованный тип в позиции параметра это *ast.Ident, а не литеральный
//     *ast.ChanType; именование канала — осознанное решение;
//   - срезы/массивы каналов ([]chan T) — это *ast.ArrayType, а не прямой
//     параметр-канал.
//
// LoadMode — Syntax: достаточно AST (*ast.ChanType с Dir == SEND|RECV в
// позиции параметра), типовая информация не требуется.
// Сгенерированный код (ast.IsGenerated) пропускается.
// Точечное отключение — стандартный //nolint:gidchandir.
package chandir

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-189"

// Analyzer — правило GID-189: параметры-каналы в сигнатурах указывают направление (<-chan/chan<-).
var Analyzer = &analysis.Analyzer{
	Name: "gidchandir",
	Doc:  ruleID + ": параметры-каналы в сигнатурах указывают направление (<-chan/chan<-)",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				checkParams(pass, node.Type)
			case *ast.FuncLit:
				checkParams(pass, node.Type)
			}
			return true
		})
	}
	return nil, nil
}

// checkParams проверяет список параметров сигнатуры: матчим только
// прямой параметр литерального типа chan T без направления (Dir == SEND|RECV).
// Возвращаемые значения (ft.Results) намеренно не проверяются.
func checkParams(pass *analysis.Pass, ft *ast.FuncType) {
	if ft == nil || ft.Params == nil {
		return
	}
	for _, field := range ft.Params.List {
		ch, ok := field.Type.(*ast.ChanType)
		if !ok {
			continue // именованный тип, []chan T, обычный тип — не наш случай.
		}
		if ch.Dir != (ast.SEND | ast.RECV) {
			continue // <-chan T или chan<- T — направление уже указано.
		}
		name := paramName(field)
		pass.Reportf(field.Type.Pos(),
			"%s: параметр-канал %s двунаправленный — укажите направление "+
				"(<-chan для чтения, chan<- для записи)",
			ruleID, name)
	}
}

// paramName возвращает имя параметра для диагностики. Для безымянного
// параметра (редкий случай в сигнатурах) — обобщённое описание.
func paramName(field *ast.Field) string {
	if len(field.Names) == 0 {
		return "без имени"
	}
	names := make([]string, 0, len(field.Names))
	for _, n := range field.Names {
		names = append(names, n.Name)
	}
	return strings.Join(names, ", ")
}
