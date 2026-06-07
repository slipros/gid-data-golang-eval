// Позитив + граница: snapshot.go — НЕ файл для ошибок.
// Объявление error-переменной здесь нарушает GID-169, а обычные
// (не-error) переменные и собственные error-типы — допустимы.
package model

import "github.com/pkg/errors"

// --- Позитив: error-var в неположенном файле ---

var ErrSnapshotConflict = errors.New("snapshot conflict") // want `GID-169: error "ErrSnapshotConflict" is declared in snapshot\.go\. Fix: keep layer errors in error\.go`

// --- Граница: тип, реализующий error через указатель ---

// ValidationError реализует error НА УКАЗАТЕЛЕ.
type ValidationError struct{ Field string }

func (e *ValidationError) Error() string { return e.Field }

// errValidation имеет тип *ValidationError → реализует error → нарушение.
var errValidation = &ValidationError{} // want `GID-169: error "errValidation" is declared in snapshot\.go\. Fix: keep layer errors in error\.go`

// errValidationValue имеет тип ValidationError (значение): метод Error
// объявлен на указателе, значение error НЕ реализует → не нарушение.
var errValidationValue = ValidationError{}

// --- Граница: не-error package-level переменные — вне scope правила ---

var DefaultLimit = 100

var snapshotName = "snapshot"

type Snapshot struct{ ID string }
