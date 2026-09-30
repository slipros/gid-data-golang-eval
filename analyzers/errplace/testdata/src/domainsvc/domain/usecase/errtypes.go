// Eval of GID-144 for error TYPES declared in /domain/usecase.
package usecase

type RetryError struct{ attempts int } // want `GID-144: error type "RetryError" is declared in "domainsvc/domain/usecase"\. Fix: declare the type in /domain/model, this layer only constructs it`

func (r RetryError) Error() string { return "retry" }
