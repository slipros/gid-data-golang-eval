// Eval of GID-144 for /domain/usecase: usecase also returns only model errors.
package usecase

import "fmt"

type Upload struct{}

// Positive: fmt.Errorf — error creation.
func (u *Upload) bad(id string) error {
	return fmt.Errorf("upload %s failed", id) // want `GID-144: creating an error via fmt\.Errorf is forbidden\. Fix: exchange it for an error from /domain/model`
}

// Negative (boundary): fmt.Sprintf is not an error constructor.
func (u *Upload) good(id string) string {
	return fmt.Sprintf("upload %s", id)
}
