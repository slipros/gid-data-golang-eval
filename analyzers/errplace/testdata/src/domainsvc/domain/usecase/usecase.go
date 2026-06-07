// Eval GID-144 для /domain/usecase: usecase тоже возвращает только model-ошибки.
package usecase

import "fmt"

type Upload struct{}

// Позитив: fmt.Errorf — создание ошибки.
func (u *Upload) bad(id string) error {
	return fmt.Errorf("upload %s failed", id) // want `GID-144: creating an error via fmt\.Errorf is forbidden\. Fix: exchange it for an error from /domain/model`
}

// Негатив (граница): fmt.Sprintf — не конструктор ошибок.
func (u *Upload) good(id string) string {
	return fmt.Sprintf("upload %s", id)
}
