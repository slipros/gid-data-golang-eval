// Package logfields implements rule GID-253 (slug log-fields-batch, linter
// gidlogfields): a logger chain sets its fields in one call.
//
// logrus offers two ways to attach fields to an entry: WithField for a single
// pair and WithFields for a batch. A chain that repeats WithField allocates a
// new *logrus.Entry per call and buries the payload in the call chain, where
// the reader has to collect the pairs by eye:
//
//	e.logger.
//		WithContext(ctx).
//		WithError(err).
//		WithField("offset", offset).
//		WithField("target_topic", e.fallbackTargetTopic(int(level))).
//		WithField("fallback_level", int(level)).
//		WithField("fallback_retry", int(retry)).
//		Error("send fallback message")
//
// The fields belong together, so they are written together:
//
//	e.logger.
//		WithContext(ctx).
//		WithError(err).
//		WithFields(logrus.Fields{
//			"offset":         offset,
//			"target_topic":   e.fallbackTargetTopic(int(level)),
//			"fallback_level": int(level),
//			"fallback_retry": int(retry),
//		}).
//		Error("send fallback message")
//
// Scope: a chain of logrus method calls (internal/lgr, KindLogrus), with or
// without a terminal call — logger := log.WithField(a).WithField(b) counts too.
// Two or more field calls in one chain are reported, whichever mix of
// WithField and WithFields they are; WithContext/WithError are not field calls
// and never count. A single WithField is exactly what the method is for and is
// left alone, and fields attached through separate statements (an intermediate
// entry variable) are outside a chain, hence outside the rule.
//
// slog is deliberately not covered: its With takes variadic pairs, so the
// batch form is the only form — there is nothing to report.
//
// Escape hatch: //nolint:gidlogfields on the chain.
package logfields

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/lgr"
)

const ruleID = "GID-253"

// fieldMethods — logrus methods that attach payload fields to an entry.
// WithContext/WithError are excluded on purpose: they carry the context and
// the error, which have their own dedicated methods (GID-155), not fields.
var fieldMethods = map[string]struct{}{
	"WithField": {}, "WithFields": {},
}

// Analyzer — rule GID-253: fields of one logger chain are set in a single
// WithFields call.
var Analyzer = &analysis.Analyzer{
	Name: "gidlogfields",
	Doc: ruleID + ": a logger chain sets its fields one by one — repeated WithField allocates an entry per " +
		`call and scatters the payload. Fix: pass them in one WithFields(logrus.Fields{"offset": offset, ` +
		`"fallback_level": level})`,
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		// A chain is checked once, from its outermost call; the inner calls are
		// consumed by that walk and skipped when Inspect reaches them.
		consumed := make(map[*ast.CallExpr]struct{})
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if _, seen := consumed[call]; seen {
				return true
			}
			checkChain(pass, call, consumed)
			return true
		})
	}
	return nil, nil
}

// checkChain walks the logrus chain the call ends with, from the outermost
// call inward, and reports the chain when it sets fields more than once.
func checkChain(pass *analysis.Pass, call *ast.CallExpr, consumed map[*ast.CallExpr]struct{}) {
	// fields hold the field calls in reverse source order (outermost first).
	var fields []*ast.SelectorExpr
	for cur := ast.Expr(call); ; {
		c, ok := cur.(*ast.CallExpr)
		if !ok {
			break
		}
		sel, ok := c.Fun.(*ast.SelectorExpr)
		if !ok || lgr.MethodKind(pass, sel) != lgr.KindLogrus {
			break
		}
		consumed[c] = struct{}{}
		if _, isField := fieldMethods[sel.Sel.Name]; isField {
			fields = append(fields, sel)
		}
		cur = sel.X
	}
	// minFieldCalls — the number of field calls that turns a chain into a
	// violation: two pairs already fit a single logrus.Fields literal.
	const minFieldCalls = 2
	if len(fields) < minFieldCalls {
		return
	}
	// Report on the first field call in source order — the chain is rewritten
	// starting there, and the position does not drift with the field count.
	pass.Reportf(fields[len(fields)-1].Sel.Pos(),
		"%s: a logger chain sets its fields in %d separate calls — they belong in one. "+
			`Fix: replace them with a single WithFields(logrus.Fields{"offset": offset, "fallback_level": level})`,
		ruleID, len(fields))
}
