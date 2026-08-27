// Eval for GID-271 with settings.exclude: this package IS a violation (one
// consumer), but the run that loads it excludes the port file by name, so
// analysistest expects zero diagnostics (TestExclude in analyzer_test.go).
package portfile

// Loader is consumed by the single Worker struct — a violation unless the
// port file is excluded by settings.exclude ("ports.go" or "Loader").
type Loader interface {
	Load(id string) error
}
