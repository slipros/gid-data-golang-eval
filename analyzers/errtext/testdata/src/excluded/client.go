// Eval of GID-256 settings.exclude: the flattening below is the same violation
// as in svc/client.go, cleared only because the method is on the exclusion
// list ("Client.legacyConfirm"). Under the default settings this file produces
// a diagnostic — TestExclude is what proves the setting drives it.
package excluded

import "github.com/pkg/errors"

// ErrServerError is the sentinel flattened into below.
var ErrServerError = errors.New("server error")

// Client owns the excluded method.
type Client struct{}

func (c *Client) legacyConfirm(err error) error {
	return errors.WithMessage(ErrServerError, err.Error())
}
