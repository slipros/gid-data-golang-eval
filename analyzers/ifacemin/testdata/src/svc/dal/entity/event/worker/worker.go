// Eval GID-197 boundary: an "event" segment nested below another layer
// (dal/entity/event/worker) must NOT be classified as the /event/** layer.
// pathseg.Contains would match "event" anywhere in the path, wrongly putting
// this package in scope; the anchored pathseg.HasLayer requires "event" to
// be the leading segment after the module root, so this package is out of
// scope and the unused Stop method below is not flagged.
package worker

type Runner interface {
	Run()
	Stop() // unused method — would be flagged if this package were in scope
}

type worker struct {
	runner Runner
}

func (w *worker) run() {
	w.runner.Run()
}
