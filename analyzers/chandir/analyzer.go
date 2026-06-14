// Package chandir implements rule GID-189 (Google: channel direction):
// a function/method parameter typed as a bidirectional channel chan T must
// declare a direction — <-chan T for receiving or chan<- T for sending.
// A bidirectional channel in a signature hides the caller's intent and
// allows erroneous usage (reading from a channel that should only be
// filled, and vice versa).
//
// What is matched:
//   - a function parameter func f(ch chan int);
//   - a method parameter func (s S) m(ch chan int);
//   - a function literal parameter func(ch chan int) { ... }.
//     We match specifically a literal chan T in parameter position.
//
// What is NOT matched:
//   - directional channels <-chan T / chan<- T — the direction is already declared;
//   - return values (func f() chan T) — the channel owner sometimes needs a
//     bidirectional one, left to review;
//   - struct fields (chan T) — the direction is set when passing to a function;
//   - local variables (var ch chan T) — that is channel creation;
//   - a parameter with a named channel type (type Pipe chan int; func f(p Pipe)) —
//     a named type in parameter position is an *ast.Ident, not a literal
//     *ast.ChanType; naming a channel is a deliberate decision;
//   - slices/arrays of channels ([]chan T) — that is an *ast.ArrayType, not a
//     direct channel parameter.
//
// LoadMode — Syntax: the AST is enough (*ast.ChanType with Dir == SEND|RECV in
// parameter position), type information is not required.
// Generated code (ast.IsGenerated) is skipped.
// Targeted opt-out — the standard //nolint:gidchandir.
package chandir

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-189"

// Analyzer — rule GID-189: channel parameters in signatures declare a direction (<-chan/chan<-).
var Analyzer = &analysis.Analyzer{
	Name: "gidchandir",
	Doc:  ruleID + ": channel parameters must declare a direction (<-chan/chan<-). Fix: use <-chan to receive or chan<- to send.",
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

// checkParams checks the signature's parameter list: we match only a
// direct parameter of the literal type chan T without a direction (Dir == SEND|RECV).
// Return values (ft.Results) are deliberately not checked.
func checkParams(pass *analysis.Pass, ft *ast.FuncType) {
	if ft == nil || ft.Params == nil {
		return
	}
	for _, field := range ft.Params.List {
		ch, ok := field.Type.(*ast.ChanType)
		if !ok {
			continue // a named type, []chan T, an ordinary type — not our case.
		}
		if ch.Dir != (ast.SEND | ast.RECV) {
			continue // <-chan T or chan<- T — the direction is already declared.
		}
		name := paramName(field)
		pass.Reportf(field.Type.Pos(),
			"%s: channel parameter %s is bidirectional. "+
				"Fix: declare a direction, <-chan to receive or chan<- to send.",
			ruleID, name)
	}
}

// paramName returns the parameter name for the diagnostic. For an unnamed
// parameter (a rare case in signatures) — a generic description.
func paramName(field *ast.Field) string {
	if len(field.Names) == 0 {
		return "unnamed"
	}
	names := make([]string, 0, len(field.Names))
	for _, n := range field.Names {
		names = append(names, n.Name)
	}
	return strings.Join(names, ", ")
}
