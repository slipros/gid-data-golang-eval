// Неприменимость GID-176 (часть 1): не граничный слой — pass-through допустим.
package util

type Worker struct{}

func (w *Worker) call() error { return nil }

func (w *Worker) passThrough() error {
	err := w.call()
	return err // ok: не граница (нет client / dal/repository в пути)
}
