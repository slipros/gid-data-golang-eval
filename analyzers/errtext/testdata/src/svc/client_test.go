package svc

import "github.com/pkg/errors"

// --- Non-applicability: a test lives in the same package (GID-250) and
// legitimately rebuilds an expected error from a text — it only has to LOOK
// like the production one, and nobody branches on its chain downstream. ---

func wantErr(err error) error {
	return errors.New(err.Error())
}
