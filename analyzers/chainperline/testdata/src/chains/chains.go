// Eval for GID-196 (chainperline).
package chains

import (
	"strings"

	"github.com/sirupsen/logrus"

	"chains/sub"
)

// --- Positive: a chain of 2 calls on a single line ---

func bad() string {
	return strings.NewReplacer("a", "b").Replace("aa") // want `GID-196: a chain of 2 calls must put one call per line, including the first\. Fix: break each \.Method\(\) onto its own line\.`
}

// --- Positive: the first call on the base's line ---

func partial() string {
	return strings.NewReplacer("a", "b"). // want `GID-196: a chain of 2 calls must put one call per line, including the first\. Fix: break each \.Method\(\) onto its own line\.`
						Replace("aa")
}

// --- Positive: a chain through an intermediate field ---

type job struct{}

func (j job) name() string { return "job" }

type repo struct{}

func (r repo) job() job { return job{} }

type svc struct{ r repo }

func fieldHop(s svc) string {
	return s.r.job().name() // want `GID-196: a chain of 2 calls must put one call per line, including the first\. Fix: break each \.Method\(\) onto its own line\.`
}

// --- Positive: two links on one line inside a multi-line chain ---

func twoOnOneLine(s svc) string {
	return s.r.
		job().name() // want `GID-196: a chain of 2 calls must put one call per line, including the first\. Fix: break each \.Method\(\) onto its own line\.`
}

// --- Negative: each call on its own line, including the first ---

func good() string {
	return strings.
		NewReplacer("a", "b").
		Replace("aa")
}

// --- Negative: a single inline call ---

func single() string {
	return strings.ToUpper("x")
}

// --- Edge: a nested call is not a chain; the inner chain is caught ---

func nested() string {
	return strings.ToUpper(strings.NewReplacer("a", "b").Replace("aa")) // want `GID-196: a chain of 2 calls must put one call per line, including the first\. Fix: break each \.Method\(\) onto its own line\.`
}

// --- Edge: a conversion via a selector is not a link ---

func conv(v string) string {
	return sub.Code(v).Upper()
}

// --- Edge: a call on a function's result — the function counts as the base ---

func factory() job { return job{} }

func fromFactory() string {
	return factory().name()
}

// --- Not applicable: a logrus chain is the domain of GID-156 ---

func logIt(l *logrus.Logger) {
	l.WithField("a", 1).Info("x")
}
