// Inapplicability of GID-176 (part 1): not a boundary layer — pass-through is acceptable.
package util

type Worker struct{}

func (w *Worker) call() error { return nil }

func (w *Worker) passThrough() error {
	err := w.call()
	return err // ok: not a boundary (no client / dal/repository in the path)
}
